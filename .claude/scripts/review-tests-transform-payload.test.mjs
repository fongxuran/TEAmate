import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

import { createPayload, parseUnifiedPatch } from "../transform-payload.mjs";

const fixtureDir = new URL("./fixtures/", import.meta.url);
const patchText = fs.readFileSync(new URL("pr.patch", fixtureDir), "utf-8");
const baseConsolidated = JSON.parse(
  fs.readFileSync(new URL("consolidated-review.json", fixtureDir), "utf-8"),
);
const filesMap = parseUnifiedPatch(patchText);

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function makeFinding(overrides = {}) {
  const finding = clone(baseConsolidated.consolidated_review[0]);
  return { ...finding, ...overrides };
}

function makePayload(findings, summaryOverrides = {}) {
  const consolidated = clone(baseConsolidated);
  consolidated.consolidated_review = findings;
  consolidated.summary = {
    ...consolidated.summary,
    total_issues: findings.length,
    ...summaryOverrides,
  };
  return createPayload(consolidated, filesMap, "deadbeef", 2000);
}

test("uses resolved_anchor when valid", () => {
  const payload = makePayload([
    makeFinding({
      line_end: 3,
      anchor_text: "marker = true",
      resolved_anchor: { side: "RIGHT", line: 22 },
      title: "Resolved anchor should win",
    }),
  ]);

  assert.equal(payload.comments.length, 1);
  assert.equal(payload.comments[0].side, "RIGHT");
  assert.equal(payload.comments[0].line, 22);
});

test("falls back when resolved_anchor is invalid", () => {
  const payload = makePayload([
    makeFinding({
      resolved_anchor: { side: "RIGHT", line: 999 },
      anchor_text: "marker = true",
      title: "Fallback to anchor text",
    }),
  ]);

  assert.equal(payload.comments.length, 1);
  assert.equal(payload.comments[0].line, 3);
});

test("skips inline comments for inline_eligible=false", () => {
  const payload = makePayload([
    makeFinding({
      inline_eligible: false,
      title: "Summary only finding",
    }),
  ]);

  assert.equal(payload.comments.length, 0);
});

test("supports LEFT-side anchors", () => {
  const payload = makePayload([
    makeFinding({
      resolved_anchor: { side: "LEFT", line: 2 },
      title: "Deleted-line finding",
    }),
  ]);

  assert.equal(payload.comments.length, 1);
  assert.equal(payload.comments[0].side, "LEFT");
  assert.equal(payload.comments[0].line, 2);
});

test("keeps multiline when same hunk and side", () => {
  const payload = makePayload([
    makeFinding({
      resolved_anchor: {
        side: "RIGHT",
        line: 23,
        start_side: "RIGHT",
        start_line: 22,
      },
      title: "Valid multiline",
    }),
  ]);

  assert.equal(payload.comments.length, 1);
  assert.equal(payload.comments[0].line, 23);
  assert.equal(payload.comments[0].start_side, "RIGHT");
  assert.equal(payload.comments[0].start_line, 22);
});

test("degrades multiline to single-line when range crosses hunks", () => {
  const payload = makePayload([
    makeFinding({
      resolved_anchor: {
        side: "RIGHT",
        line: 23,
        start_side: "RIGHT",
        start_line: 2,
      },
      title: "Cross-hunk multiline",
    }),
  ]);

  assert.equal(payload.comments.length, 1);
  assert.equal(payload.comments[0].line, 23);
  assert.equal("start_line" in payload.comments[0], false);
  assert.equal("start_side" in payload.comments[0], false);
});

test("deduplicates inline comments by path/side/line/title", () => {
  const first = makeFinding({
    resolved_anchor: { side: "RIGHT", line: 2 },
    title: "Duplicate issue",
  });
  const second = makeFinding({
    resolved_anchor: { side: "RIGHT", line: 2 },
    title: "  Duplicate issue  ",
    description: "Same issue from another reviewer",
  });

  const payload = makePayload([first, second]);

  assert.equal(payload.comments.length, 1);
});

test("renders nitpick marker and optional summary", () => {
  const payload = makePayload(
    [
      makeFinding({
        priority: "P3",
        is_required: false,
        file_path: "src/bar.js",
        line_start: 0,
        line_end: 6,
        anchor_text: "const sum = a + b",
        symbol: "add",
        resolved_anchor: { side: "RIGHT", line: 6 },
        title: "Rename sum variable",
        description: "Minor readability tweak.",
      }),
    ],
    {
      required_issues: 0,
      optional_issues: 1,
      by_priority: { P0: 0, P1: 0, P2: 0, P3: 1 },
    },
  );

  assert.equal(payload.comments.length, 1);
  assert.match(payload.comments[0].body, /\[nitpick\] Rename sum variable/i);
  assert.match(payload.comments[0].body, /Optional: this is not required to fix/i);
  assert.match(payload.body, /Optional \(including nitpicks\): 1/);
});
