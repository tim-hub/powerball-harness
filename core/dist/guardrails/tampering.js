/**
 * core/src/guardrails/tampering.ts
 * Test tampering detection engine
 *
 * All patterns from posttooluse-tampering-detector.sh ported to TypeScript.
 * After test files or CI configuration are modified by Write / Edit / MultiEdit tools,
 * detects tampering patterns and returns warnings (does not block).
 */
// ============================================================
// File type detection
// ============================================================
const TEST_FILE_PATTERNS = [
    /\.test\.[jt]sx?$/,
    /\.spec\.[jt]sx?$/,
    /\.test\.py$/,
    /test_[^/]+\.py$/,
    /[^/]+_test\.py$/,
    /\.test\.go$/,
    /[^/]+_test\.go$/,
    /\/__tests__\//,
    /\/tests\//,
];
const CONFIG_FILE_PATTERNS = [
    /(?:^|\/)\.eslintrc(?:\.[^/]+)?$/,
    /(?:^|\/)eslint\.config\.[^/]+$/,
    /(?:^|\/)\.prettierrc(?:\.[^/]+)?$/,
    /(?:^|\/)prettier\.config\.[^/]+$/,
    /(?:^|\/)tsconfig(?:\.[^/]+)?\.json$/,
    /(?:^|\/)biome\.json$/,
    /(?:^|\/)\.stylelintrc(?:\.[^/]+)?$/,
    /(?:^|\/)(?:jest|vitest)\.config\.[^/]+$/,
    /\.github\/workflows\/[^/]+\.ya?ml$/,
    /(?:^|\/)\.gitlab-ci\.ya?ml$/,
    /(?:^|\/)Jenkinsfile$/,
];
function isTestFile(filePath) {
    return TEST_FILE_PATTERNS.some((p) => p.test(filePath));
}
function isConfigFile(filePath) {
    return CONFIG_FILE_PATTERNS.some((p) => p.test(filePath));
}
const TAMPERING_PATTERNS = [
    // --- Test skipping ---
    {
        id: "T01:it-skip",
        description: "Test skipping via it.skip / describe.skip",
        pattern: /(?:it|test|describe|context)\.skip\s*\(/,
        testFileOnly: true,
    },
    {
        id: "T02:xit-xdescribe",
        description: "Test disabling via xit / xdescribe",
        pattern: /\b(?:xit|xtest|xdescribe)\s*\(/,
        testFileOnly: true,
    },
    {
        id: "T03:pytest-skip",
        description: "Test skipping via pytest.mark.skip",
        pattern: /@pytest\.mark\.(?:skip|xfail)\b/,
        testFileOnly: true,
    },
    {
        id: "T04:go-skip",
        description: "Test skipping via t.Skip()",
        pattern: /\bt\.Skip(?:f|Now)?\s*\(/,
        testFileOnly: true,
    },
    // --- Assertion removal/tampering ---
    {
        id: "T05:expect-removed",
        description: "Possible removal of expect / assert (commented out)",
        pattern: /\/\/\s*expect\s*\(/,
        testFileOnly: true,
    },
    {
        id: "T06:assert-commented",
        description: "Assert call commented out",
        pattern: /\/\/\s*assert(?:Equal|NotEqual|True|False|Nil|Error)?\s*\(/,
        testFileOnly: true,
    },
    {
        id: "T07:todo-assert",
        description: "Assertion replaced by a TODO comment",
        pattern: /\/\/\s*TODO.*assert|\/\/\s*TODO.*expect/i,
        testFileOnly: true,
    },
    // --- ESLint disabling (test and CI config) ---
    {
        id: "T08:eslint-disable",
        description: "Lint rule disabling via eslint-disable",
        // Supports both // eslint-disable and /* eslint-disable */ formats
        pattern: /(?:\/\/\s*eslint-disable(?:-next-line|-line)?(?:\s+[^\n]+)?$|\/\*\s*eslint-disable\b[^*]*\*\/)/m,
        testFileOnly: false,
    },
    // --- CI workflow tampering ---
    {
        id: "T09:ci-continue-on-error",
        description: "CI failure suppression via continue-on-error: true",
        pattern: /continue-on-error\s*:\s*true/,
        testFileOnly: false,
    },
    {
        id: "T10:ci-if-always",
        description: "CI step forced execution via if: always()",
        pattern: /if\s*:\s*always\s*\(\s*\)/,
        testFileOnly: false,
    },
    // --- Hardcoded expected values ---
    {
        id: "T11:hardcoded-answer",
        description: "Hardcoded test expected values (dictionary lookup)",
        pattern: /answers?_for_tests?\s*=\s*\{/,
        testFileOnly: true,
    },
    {
        id: "T12:return-hardcoded",
        description: "Pattern returning test case values directly",
        pattern: /return\s+(?:"[^"]*"|'[^']*'|\d+)\s*;\s*\/\/.*(?:test|spec|expect)/i,
        testFileOnly: true,
    },
];
/**
 * Search for tampering patterns in text (new_string or content).
 */
function detectTampering(text, isTest) {
    const warnings = [];
    for (const p of TAMPERING_PATTERNS) {
        if (p.testFileOnly && !isTest)
            continue;
        const match = p.pattern.exec(text);
        if (match !== null) {
            warnings.push({
                patternId: p.id,
                description: p.description,
                matchedText: match[0].slice(0, 120),
            });
        }
    }
    return warnings;
}
/**
 * Extract file path and changed text from HookInput.
 */
function extractTargets(input) {
    const toolInput = input.tool_input;
    const filePath = toolInput["file_path"];
    if (typeof filePath !== "string" || filePath.length === 0)
        return null;
    // Write: content field
    // Edit: new_string field
    const changedText = typeof toolInput["content"] === "string"
        ? toolInput["content"]
        : typeof toolInput["new_string"] === "string"
            ? toolInput["new_string"]
            : null;
    if (changedText === null)
        return null;
    return { filePath, changedText };
}
// ============================================================
// Export: PostToolUse entry point
// ============================================================
/**
 * Detect test tampering in a PostToolUse hook and return warnings.
 * Even when tampering is detected, the decision is "approve" (does not block).
 * Warnings are passed to Claude as systemMessage.
 */
export function detectTestTampering(input) {
    // Only target Write / Edit / MultiEdit
    if (!["Write", "Edit", "MultiEdit"].includes(input.tool_name)) {
        return { decision: "approve" };
    }
    const targets = extractTargets(input);
    if (targets === null)
        return { decision: "approve" };
    const { filePath, changedText } = targets;
    const isTest = isTestFile(filePath);
    const isConfig = isConfigFile(filePath);
    if (!isTest && !isConfig)
        return { decision: "approve" };
    const warnings = detectTampering(changedText, isTest);
    if (warnings.length === 0)
        return { decision: "approve" };
    const fileType = isTest ? "test file" : "CI/config file";
    const warningLines = warnings
        .map((w) => `- [${w.patternId}] ${w.description}\n  Detected: ${w.matchedText}`)
        .join("\n");
    const systemMessage = `[Harness v3] Test tampering detection warning\n\n` +
        `Suspicious patterns detected in ${fileType} \`${filePath}\`:\n\n` +
        warningLines +
        `\n\n[Please verify]\n` +
        `Ensure this change does not intentionally disable tests or degrade implementation quality.\n` +
        `If determined to be tampering, revert the change.`;
    return {
        decision: "approve",
        systemMessage,
    };
}
//# sourceMappingURL=tampering.js.map