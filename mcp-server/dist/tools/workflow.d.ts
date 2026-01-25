/**
 * Workflow Tools
 *
 * Core Harness workflow operations accessible via MCP.
 * Enables Plan → Work → Review cycle from any MCP client.
 */
import { type Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const workflowTools: Tool[];
export declare function handleWorkflowTool(name: string, args: Record<string, unknown> | undefined): Promise<{
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}>;
//# sourceMappingURL=workflow.d.ts.map