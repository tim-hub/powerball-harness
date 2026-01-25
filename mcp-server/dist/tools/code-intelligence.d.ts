/**
 * Code Intelligence Tools
 *
 * AST-Grep and LSP integration for enhanced code analysis.
 * These tools enable structural code search and semantic analysis.
 *
 * Requires external setup:
 * - AST-Grep: `brew install ast-grep` or `npm install -g @ast-grep/cli`
 * - LSP: Language-specific servers (typescript-language-server, etc.)
 *
 * Run `/dev-tools-setup` to install and configure these tools.
 */
import { type Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const codeIntelligenceTools: Tool[];
export declare function handleCodeIntelligenceTool(name: string, args: Record<string, unknown> | undefined): Promise<{
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}>;
//# sourceMappingURL=code-intelligence.d.ts.map