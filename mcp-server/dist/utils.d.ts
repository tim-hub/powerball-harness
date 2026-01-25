/**
 * Shared Utilities for Harness MCP Server
 *
 * Common functions used across multiple tool modules.
 */
import { exec } from "child_process";
export declare const execAsync: typeof exec.__promisify__;
/** Session is considered stale after this many seconds (1 hour) */
export declare const STALE_THRESHOLD_SECONDS = 3600;
/** Maximum number of broadcast messages to retain */
export declare const MAX_BROADCAST_MESSAGES = 100;
/** Directory for session state files */
export declare const SESSIONS_DIR = ".claude/sessions";
/** Active sessions file path */
export declare const ACTIVE_SESSIONS_FILE = ".claude/sessions/active.json";
/** Broadcast messages file path (Markdown format for CLI compatibility) */
export declare const BROADCAST_FILE = ".claude/sessions/broadcast.md";
/**
 * Find the project root by looking for common markers.
 * Traverses up the directory tree until a marker is found.
 * Compatible with both Unix and Windows file systems.
 *
 * @returns The project root path, or current working directory if not found
 */
export declare function getProjectRoot(): string;
/**
 * Ensure a directory exists, creating it if necessary.
 *
 * @param dirPath - The directory path to ensure exists
 */
export declare function ensureDir(dirPath: string): void;
/**
 * Safely parse JSON from a file with error logging.
 *
 * @param filePath - Path to the JSON file
 * @param defaultValue - Default value if file doesn't exist or parse fails
 * @returns Parsed JSON or default value
 */
export declare function safeReadJSON<T>(filePath: string, defaultValue: T): T;
/**
 * Safely write JSON to a file with error logging.
 *
 * @param filePath - Path to write the JSON file
 * @param data - Data to serialize and write
 * @returns true if successful, false otherwise
 */
export declare function safeWriteJSON<T>(filePath: string, data: T): boolean;
/**
 * Validate that a path is safe for use in shell commands.
 * Prevents command injection by rejecting paths with dangerous characters.
 *
 * @param inputPath - The path to validate
 * @returns true if the path is safe, false otherwise
 */
export declare function isValidPath(inputPath: string): boolean;
/**
 * Get list of recently changed files using git diff (async).
 * Validates the base path to prevent command injection.
 *
 * @param basePath - The repository path (validated for safety)
 * @returns Array of changed file paths
 */
export declare function getRecentChangesAsync(basePath?: string): Promise<string[]>;
/**
 * Format a duration in seconds to a human-readable string.
 *
 * @param seconds - Duration in seconds
 * @returns Human-readable duration (e.g., "5m ago", "2h ago")
 */
export declare function formatTimeAgo(seconds: number): string;
//# sourceMappingURL=utils.d.ts.map