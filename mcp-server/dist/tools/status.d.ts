/**
 * Status Tools
 *
 * Project status and synchronization tools.
 */
import { type Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const statusTools: Tool[];
export declare function handleStatusTool(name: string, args: Record<string, unknown> | undefined): Promise<{
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}>;
//# sourceMappingURL=status.d.ts.map