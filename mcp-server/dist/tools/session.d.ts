/**
 * Session Communication Tools
 *
 * Enables inter-session messaging across different AI clients.
 * Works with Claude Code, Codex, and other MCP-compatible clients.
 */
import { type Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const sessionTools: Tool[];
export declare function handleSessionTool(name: string, args: Record<string, unknown> | undefined): Promise<{
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}>;
//# sourceMappingURL=session.d.ts.map