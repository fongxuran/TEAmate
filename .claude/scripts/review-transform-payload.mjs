#!/usr/bin/env node
/**
 * Generate GitHub "Create a review" payload with inline comments from:
 * - consolidated-review.json
 * - PR unified patch (gh pr diff <PR> --patch)
 *
 * Usage:
 *   node transform-payload.mjs \
 *     --consolidated consolidated-review.json \
 *     --patch pr.patch \
 *     --commit 73db78f8541f2a31b65a5425030087f1d6e992f4 \
 *     --out github-review-payload.json
 *
 * Optional:
 *   --maxDiffChars 2000
 */

import fs from "node:fs";
import process from "node:process";
import { pathToFileURL } from "node:url";

function die(msg) {
  console.error(msg);
  process.exit(1);
}

function parseArgs(argv) {
  const args = {
    consolidated: null,
    patch: null,
    commit: null,
    out: "github-review-payload.json",
    maxDiffChars: 2000,
  };

  const a = argv.slice(2);
  for (let i = 0; i < a.length; i++) {
    const k = a[i];
    const v = a[i + 1];
    if (k === "--consolidated") args.consolidated = v, i++;
    else if (k === "--patch") args.patch = v, i++;
    else if (k === "--commit") args.commit = v, i++;
    else if (k === "--out") args.out = v, i++;
    else if (k === "--maxDiffChars") args.maxDiffChars = Number(v), i++;
    else die(`Unknown arg: ${k}`);
  }

  if (!args.consolidated) die("Missing --consolidated <file>");
  if (!args.patch) die("Missing --patch <file>");
  if (!args.commit) die("Missing --commit <sha>");
  if (!Number.isFinite(args.maxDiffChars) || args.maxDiffChars < 200) {
    die("Invalid --maxDiffChars");
  }

  return args;
}

function normalizeNewlines(s) {
  return String(s || "").replace(/\r\n/g, "\n");
}

function normalizeForMatch(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}

function trimBlock(s, maxChars) {
  const t = String(s || "").trim();
  if (!t) return "";
  if (t.length <= maxChars) return t;
  return `${t.slice(0, maxChars)}\n...\n`;
}

/**
 * Parse unified diff produced by `gh pr diff --patch`.
 * Builds side-aware line maps with hunk ids for multiline validation.
 */
function parseUnifiedPatch(patchText) {
  const lines = normalizeNewlines(patchText).split("\n");
  const files = new Map();

  let curPath = null;
  let oldLine = null;
  let newLine = null;
  let hunkId = 0;

  const ensureFile = (p) => {
    if (!files.has(p)) {
      files.set(p, {
        rightAll: [],
        rightChanged: [],
        leftAll: [],
        leftChanged: [],
        rightLineMap: new Map(),
        leftLineMap: new Map(),
      });
    }
    return files.get(p);
  };

  const pushEntry = (file, side, line, text, changed, hunk) => {
    const entry = { line, text, isChanged: changed, hunkId: hunk };
    if (side === "RIGHT") {
      file.rightAll.push(entry);
      file.rightLineMap.set(line, entry);
      if (changed) file.rightChanged.push(entry);
    } else {
      file.leftAll.push(entry);
      file.leftLineMap.set(line, entry);
      if (changed) file.leftChanged.push(entry);
    }
  };

  const diffGitRe = /^diff --git a\/(.+?) b\/(.+)$/;
  const hunkRe = /^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@/;

  for (let i = 0; i < lines.length; i++) {
    const l = lines[i];

    const dg = l.match(diffGitRe);
    if (dg) {
      curPath = dg[2];
      oldLine = null;
      newLine = null;
      ensureFile(curPath);
      continue;
    }

    if (!curPath) continue;

    const h = l.match(hunkRe);
    if (h) {
      hunkId += 1;
      oldLine = Number(h[1]);
      newLine = Number(h[3]);
      continue;
    }

    if (oldLine === null || newLine === null) continue;

    const f = ensureFile(curPath);

    if (l.startsWith(" ")) {
      const text = l.slice(1);
      pushEntry(f, "RIGHT", newLine, text, false, hunkId);
      pushEntry(f, "LEFT", oldLine, text, false, hunkId);
      oldLine += 1;
      newLine += 1;
    } else if (l.startsWith("+") && !l.startsWith("+++")) {
      pushEntry(f, "RIGHT", newLine, l.slice(1), true, hunkId);
      newLine += 1;
    } else if (l.startsWith("-") && !l.startsWith("---")) {
      pushEntry(f, "LEFT", oldLine, l.slice(1), true, hunkId);
      oldLine += 1;
    }
  }

  return files;
}

function getSideData(file, side) {
  if (side === "RIGHT") {
    return {
      all: file.rightAll,
      changed: file.rightChanged,
      lineMap: file.rightLineMap,
    };
  }
  return {
    all: file.leftAll,
    changed: file.leftChanged,
    lineMap: file.leftLineMap,
  };
}

function toFiniteInt(value) {
  if (!Number.isFinite(value)) return null;
  return Number(value);
}

function lineEntry(file, side, line) {
  if (!file || !side || !Number.isFinite(line)) return null;
  const data = getSideData(file, side);
  return data.lineMap.get(Number(line)) || null;
}

function withValidatedMultiline(file, anchor) {
  if (!anchor.start_side || !Number.isFinite(anchor.start_line)) {
    return { path: anchor.path, side: anchor.side, line: anchor.line };
  }

  if (anchor.start_side !== anchor.side) {
    return { path: anchor.path, side: anchor.side, line: anchor.line };
  }

  const endEntry = lineEntry(file, anchor.side, anchor.line);
  const startEntry = lineEntry(file, anchor.side, anchor.start_line);
  if (!endEntry || !startEntry || endEntry.hunkId !== startEntry.hunkId) {
    return { path: anchor.path, side: anchor.side, line: anchor.line };
  }

  const minLine = Math.min(anchor.start_line, anchor.line);
  const maxLine = Math.max(anchor.start_line, anchor.line);
  if (minLine === maxLine) {
    return { path: anchor.path, side: anchor.side, line: maxLine };
  }

  return {
    path: anchor.path,
    side: anchor.side,
    line: maxLine,
    start_side: anchor.side,
    start_line: minLine,
  };
}

function resolveFromConsolidatedAnchor(filesMap, finding) {
  const fp = String(finding.file_path || "").trim();
  const resolved = finding.resolved_anchor;
  if (!fp || !resolved || typeof resolved !== "object") return null;

  const file = filesMap.get(fp);
  if (!file) return null;

  const side = String(resolved.side || "").toUpperCase();
  const line = toFiniteInt(resolved.line);
  if ((side !== "RIGHT" && side !== "LEFT") || line === null) return null;
  if (!lineEntry(file, side, line)) return null;

  const startSide = String(resolved.start_side || "").toUpperCase();
  const startLine = toFiniteInt(resolved.start_line);

  const anchor = { path: fp, side, line };
  if ((startSide === "RIGHT" || startSide === "LEFT") && startLine !== null) {
    anchor.start_side = startSide;
    anchor.start_line = startLine;
  }

  return withValidatedMultiline(file, anchor);
}

function findTextMatch(sideData, needle) {
  const target = String(needle || "").trim();
  if (!target) return null;

  const exact = (arr) => arr.find((x) => x.text.includes(target)) || null;
  const fuzzyNeedle = normalizeForMatch(target);
  const fuzzy = (arr) => arr.find((x) => normalizeForMatch(x.text).includes(fuzzyNeedle)) || null;

  return (
    exact(sideData.changed)
    || exact(sideData.all)
    || fuzzy(sideData.changed)
    || fuzzy(sideData.all)
    || null
  );
}

function pickWithinRange(sideData, start, end) {
  if (!Number.isFinite(start) || !Number.isFinite(end)) return null;
  const minLine = Math.min(start, end);
  const maxLine = Math.max(start, end);
  const candidates = sideData.all.filter((x) => x.line >= minLine && x.line <= maxLine);
  if (!candidates.length) return null;
  const changed = candidates.find((x) => x.isChanged);
  return changed || candidates[0];
}

function pickClosest(arr, target) {
  if (!arr.length || !Number.isFinite(target)) return null;
  let best = null;
  let bestDist = Infinity;
  for (const item of arr) {
    const dist = Math.abs(item.line - target);
    if (dist < bestDist) {
      best = item;
      bestDist = dist;
    }
  }
  return best;
}

function pickNearestChanged(sideData, target) {
  if (!sideData.changed.length) return null;
  if (!Number.isFinite(target)) return sideData.changed[0];
  return pickClosest(sideData.changed, target);
}

function buildRangeAnchor(path, side, sideData, lineStart, lineEnd, file) {
  const end = toFiniteInt(lineEnd);
  if (end === null) return null;

  const startCandidate = toFiniteInt(lineStart);
  const rangeStart = Number.isFinite(startCandidate) && startCandidate > 0 ? startCandidate : end;

  const endEntry = pickWithinRange(sideData, rangeStart, end)
    || pickClosest(sideData.all, end)
    || pickNearestChanged(sideData, end);
  if (!endEntry) return null;

  const anchor = { path, side, line: endEntry.line };

  if (Number.isFinite(startCandidate) && startCandidate > 0 && end > startCandidate) {
    const startEntry = pickWithinRange(sideData, startCandidate, end)
      || pickClosest(sideData.all, startCandidate)
      || pickNearestChanged(sideData, startCandidate);

    if (startEntry) {
      anchor.start_side = side;
      anchor.start_line = startEntry.line;
    }
  }

  return withValidatedMultiline(file, anchor);
}

function fallbackAnchorForFinding(filesMap, finding) {
  const fp = String(finding.file_path || "").trim();
  if (!fp) return null;

  const file = filesMap.get(fp);
  if (!file) return null;

  const anchorText = String(finding.anchor_text || "").trim();
  const symbol = String(finding.symbol || "").trim();
  const ls = toFiniteInt(finding.line_start);
  const le = toFiniteInt(finding.line_end);
  const target = le ?? ls;

  const trySide = (side) => {
    const sideData = getSideData(file, side);

    if (anchorText) {
      const byText = findTextMatch(sideData, anchorText);
      if (byText) return { path: fp, side, line: byText.line };
    }

    if (symbol) {
      const bySymbol = findTextMatch(sideData, symbol);
      if (bySymbol) return { path: fp, side, line: bySymbol.line };
    }

    const byRange = buildRangeAnchor(fp, side, sideData, ls, le, file);
    if (byRange) return byRange;

    const nearest = pickNearestChanged(sideData, target);
    if (nearest) return { path: fp, side, line: nearest.line };

    return null;
  };

  return trySide("RIGHT") || trySide("LEFT") || null;
}

function isInlineEligible(finding) {
  return finding.inline_eligible !== false;
}

function anchorForFinding(filesMap, finding) {
  if (!isInlineEligible(finding)) return null;

  const fromResolved = resolveFromConsolidatedAnchor(filesMap, finding);
  if (fromResolved) return fromResolved;

  return fallbackAnchorForFinding(filesMap, finding);
}

function severityHeader(priority, emojiCategory, isNitpick) {
  const cat = String(emojiCategory || "").trim();
  const map = {
    P0: "Blocking issue",
    P1: "Potential issue",
    P2: "Suggested improvement",
    P3: "Nitpick",
  };
  const base = isNitpick ? "Nitpick (optional)" : (cat || map[priority] || "Potential issue");
  return `_${base}_ **[${priority || "Unknown"}]**`;
}

function findingTitle(finding) {
  const raw = String(finding.title || "").trim() || "(Untitled finding)";
  const isNitpick = finding.priority === "P3" && finding.is_required === false;
  if (!isNitpick) return raw;
  return /^\[nitpick\]/i.test(raw) ? raw : `[nitpick] ${raw}`;
}

function detailsBlock(summary, innerMarkdown) {
  return [
    "<details>",
    `<summary>${summary}</summary>`,
    "",
    innerMarkdown.trim(),
    "",
    "</details>",
  ].join("\n");
}

function buildCommentBody(finding, maxDiffChars) {
  const isNitpick = finding.priority === "P3" && finding.is_required === false;
  const header = severityHeader(finding.priority, finding.emoji_category, isNitpick);
  const title = findingTitle(finding);
  const description = String(finding.description || "").trim();
  const impact = String(finding.impact || "").trim();
  const recommendation = String(finding.recommendation || "").trim();

  const parts = [];
  parts.push(header, "", `**${title}**`, "");

  if (isNitpick) {
    parts.push("- Optional: this is not required to fix in this PR.", "");
  }

  if (description) parts.push(description, "");

  if (impact) {
    parts.push("**Why it matters**");
    parts.push(`- ${impact}`);
    parts.push("");
  }

  if (recommendation) {
    parts.push("**Recommended fix**");
    if (/^\s*\d+[\).\]]/m.test(recommendation)) parts.push(recommendation);
    else parts.push(`- ${recommendation}`);
    parts.push("");
  }

  const codeDiff = trimBlock(finding.code_diff, maxDiffChars);
  if (codeDiff) {
    parts.push("Apply this diff:", "", "```diff", codeDiff, "```", "");
  }

  if (finding.has_suggestion && String(finding.suggestion_code || "").trim()) {
    const sug = String(finding.suggestion_code).trim();
    const inner = [
      "> IMPORTANT",
      "> Review this suggestion carefully before committing. Ensure no missing lines or indentation issues, then run relevant tests.",
      "",
      "```suggestion",
      sug,
      "```",
    ].join("\n");

    parts.push("<!-- suggestion_start -->", "");
    parts.push(detailsBlock("Committable suggestion", inner));
    parts.push("", "<!-- suggestion_end -->", "");
  }

  if (String(finding.ai_prompt || "").trim()) {
    const inner = ["```", String(finding.ai_prompt).trim(), "```"].join("\n");
    parts.push(detailsBlock("Prompt for AI agents", inner), "");
  }

  parts.push("<!-- This is an auto-generated comment by Review Orchestrator -->");
  parts.push("", "_AI Review_");

  return parts.join("\n").replace(/\n{3,}/g, "\n\n").trim();
}

function buildTopLevelBody(consolidated, findings) {
  const total = consolidated?.summary?.total_issues ?? findings.length;
  const title = `# PR Review (PENDING)${total ? ` — ${total} findings` : ""}`;

  const byPriority = consolidated?.summary?.by_priority || {
    P0: findings.filter((f) => f.priority === "P0").length,
    P1: findings.filter((f) => f.priority === "P1").length,
    P2: findings.filter((f) => f.priority === "P2").length,
    P3: findings.filter((f) => f.priority === "P3").length,
  };

  const required = consolidated?.summary?.required_issues
    ?? findings.filter((f) => f.is_required !== false).length;
  const optional = consolidated?.summary?.optional_issues
    ?? findings.filter((f) => f.is_required === false).length;

  const topRisks = Array.isArray(consolidated.top_risks) ? consolidated.top_risks : [];

  return [
    title,
    "",
    "## Summary",
    `- Findings: P0=${byPriority.P0 ?? 0}, P1=${byPriority.P1 ?? 0}, P2=${byPriority.P2 ?? 0}, P3=${byPriority.P3 ?? 0}`,
    `- Required: ${required}`,
    `- Optional (including nitpicks): ${optional}`,
    "",
    topRisks.length ? "## Top risks" : "",
    ...topRisks.map((r) => `- ${r}`),
    "",
    "_AI Review_",
  ].filter(Boolean).join("\n");
}

function createPayload(consolidated, filesMap, commit, maxDiffChars) {
  const findings = Array.isArray(consolidated.consolidated_review)
    ? consolidated.consolidated_review
    : [];

  if (!findings.length) die("consolidated_review[] is empty");

  const comments = [];
  const seen = new Set();

  for (const finding of findings) {
    if (!isInlineEligible(finding)) continue;

    const anchor = anchorForFinding(filesMap, finding);
    if (!anchor) continue;

    const titleKey = normalizeForMatch(findingTitle(finding));
    const key = `${anchor.path}|${anchor.side}|${anchor.line}|${titleKey}`;
    if (seen.has(key)) continue;
    seen.add(key);

    const comment = {
      path: anchor.path,
      body: buildCommentBody(finding, maxDiffChars),
      side: anchor.side,
      line: anchor.line,
    };

    if (anchor.start_side && Number.isFinite(anchor.start_line)) {
      comment.start_side = anchor.start_side;
      comment.start_line = anchor.start_line;
    }

    comments.push(comment);
  }

  return {
    commit_id: commit,
    body: buildTopLevelBody(consolidated, findings),
    comments,
  };
}

function main() {
  const args = parseArgs(process.argv);

  const consolidatedRaw = fs.readFileSync(args.consolidated, "utf-8");
  const consolidated = JSON.parse(consolidatedRaw);

  const patch = fs.readFileSync(args.patch, "utf-8");
  const filesMap = parseUnifiedPatch(patch);

  const payload = createPayload(consolidated, filesMap, args.commit, args.maxDiffChars);

  fs.writeFileSync(args.out, JSON.stringify(payload, null, 2) + "\n", "utf-8");
  console.log(`Wrote payload: ${args.out}`);
  console.log(`Inline comments: ${payload.comments.length}`);
  console.log(`Anchored files: ${new Set(payload.comments.map((c) => c.path)).size}`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main();
}

export {
  parseUnifiedPatch,
  anchorForFinding,
  buildCommentBody,
  buildTopLevelBody,
  createPayload,
};
