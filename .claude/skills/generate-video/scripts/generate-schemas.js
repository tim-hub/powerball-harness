#!/usr/bin/env node
/**
 * @file generate-schemas.js
 * @description Auto-generate Zod schemas from JSON Schema
 *
 * Usage:
 *   npm run generate:schemas
 *
 * Input:  schemas/*.schema.json
 * Output: src/schemas/*.ts (Zod definitions)
 */

const fs = require('fs');
const path = require('path');
const { jsonSchemaToZod } = require('json-schema-to-zod');
const deref = require('json-schema-deref-sync');

// Directory paths
const SCHEMAS_DIR = path.join(__dirname, '../schemas');
const OUTPUT_DIR = path.join(__dirname, '../src/schemas');

/**
 * Read JSON Schema files and convert them to Zod schemas
 */
function generateSchemas() {
  console.log('🔧 Starting schema generation...\n');

  // Verify schemas directory exists
  if (!fs.existsSync(SCHEMAS_DIR)) {
    console.error(`❌ Error: schemas directory not found at ${SCHEMAS_DIR}`);
    process.exit(1);
  }

  // Create output directory
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
    console.log(`✅ Created output directory: ${OUTPUT_DIR}\n`);
  }

  // Search for .schema.json files
  const schemaFiles = fs
    .readdirSync(SCHEMAS_DIR)
    .filter(file => file.endsWith('.schema.json'));

  if (schemaFiles.length === 0) {
    console.warn('⚠️  No .schema.json files found in schemas/');
    process.exit(0);
  }

  console.log(`📂 Found ${schemaFiles.length} schema file(s):\n`);

  const results = [];

  // Convert each schema file
  for (const schemaFile of schemaFiles) {
    const schemaPath = path.join(SCHEMAS_DIR, schemaFile);
    const baseName = schemaFile.replace('.schema.json', '');
    const outputPath = path.join(OUTPUT_DIR, `${baseName}.ts`);

    try {
      console.log(`  Processing: ${schemaFile}`);

      // Read JSON Schema
      const schemaContent = fs.readFileSync(schemaPath, 'utf-8');
      const schema = JSON.parse(schemaContent);

      // Resolve $ref references (pre-dereference)
      const dereferencedSchema = deref(schema, {
        baseFolder: SCHEMAS_DIR,
        failOnMissing: true,
      });

      // Convert to Zod schema
      const zodSchema = jsonSchemaToZod(dereferencedSchema, {
        module: 'esm', // ES Modules format
        name: toPascalCase(baseName), // Convert schema name to PascalCase
      });

      // Output as TypeScript file
      const outputContent = generateTypeScriptFile(baseName, zodSchema, schema);
      fs.writeFileSync(outputPath, outputContent, 'utf-8');

      console.log(`    ✅ Generated: ${path.relative(process.cwd(), outputPath)}`);
      results.push({ file: schemaFile, success: true });

    } catch (error) {
      console.error(`    ❌ Error processing ${schemaFile}:`, error.message);
      if (error.message && error.message.includes('$ref')) {
        console.error(`       Hint: Check that all $ref paths exist in ${SCHEMAS_DIR}`);
      }
      results.push({ file: schemaFile, success: false, error: error.message });
    }
  }

  // Generate index file
  generateIndexFile(schemaFiles);

  // Results summary
  console.log('\n📊 Generation Summary:');
  const successful = results.filter(r => r.success).length;
  const failed = results.filter(r => !r.success).length;

  console.log(`  ✅ Successful: ${successful}`);
  if (failed > 0) {
    console.log(`  ❌ Failed: ${failed}`);
    process.exit(1);
  }

  console.log('\n✨ Schema generation completed successfully!');
}

/**
 * Generate TypeScript file content
 */
function generateTypeScriptFile(baseName, zodSchema, originalSchema) {
  const schemaName = toPascalCase(baseName);
  const version = originalSchema.version || '1.0.0';
  const title = originalSchema.title || schemaName;
  const description = originalSchema.description || '';

  // Remove duplicate import from generated zodSchema (json-schema-to-zod adds its own import)
  let cleanedZodSchema = zodSchema
    .replace(/^import \{ z \} from ['"]zod['"];?\n?/gm, '')
    .replace(/^import \{ z \} from "zod"\n?/gm, '')
    .trim();

  // Fix variable naming: ensure export uses 'Schema' suffix
  // json-schema-to-zod generates 'export const Name = ...'
  // We need 'export const NameSchema = ...'
  const exportPattern = new RegExp(`export const ${schemaName} = `, 'g');
  if (cleanedZodSchema.match(exportPattern)) {
    cleanedZodSchema = cleanedZodSchema.replace(exportPattern, `export const ${schemaName}Schema = `);
  } else {
    // If no export found, wrap the schema
    cleanedZodSchema = `export const ${schemaName}Schema = ${cleanedZodSchema}`;
  }

  return `/**
 * @file ${baseName}.ts
 * @description Auto-generated Zod schema for ${title}
 * @version ${version}
 * @generated This file is auto-generated from schemas/${baseName}.schema.json
 *           All $ref references are resolved during generation.
 *           DO NOT EDIT MANUALLY - run \`npm run generate:schemas\` instead
 */

import { z } from 'zod';

/**
 * ${description}
 */
${cleanedZodSchema}

/**
 * Inferred TypeScript type from Zod schema
 */
export type ${schemaName} = z.infer<typeof ${schemaName}Schema>;

/**
 * Schema metadata
 */
export const ${schemaName}Meta = {
  version: '${version}',
  title: '${title}',
  description: '${description.replace(/'/g, "\\'")}',
} as const;
`;
}

/**
 * Generate index file
 */
function generateIndexFile(schemaFiles) {
  const indexPath = path.join(OUTPUT_DIR, 'index.ts');
  const exports = schemaFiles
    .map(file => {
      const baseName = file.replace('.schema.json', '');
      return `export * from './${baseName}';`;
    })
    .join('\n');

  const indexContent = `/**
 * @file index.ts
 * @description Auto-generated barrel export for all schemas
 * @generated This file is auto-generated - DO NOT EDIT MANUALLY
 */

${exports}
`;

  fs.writeFileSync(indexPath, indexContent, 'utf-8');
  console.log(`\n  ✅ Generated index: ${path.relative(process.cwd(), indexPath)}`);
}

/**
 * Convert a string to PascalCase
 */
function toPascalCase(str) {
  return str
    .split(/[-_.]/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join('');
}

// Execute
if (require.main === module) {
  try {
    generateSchemas();
  } catch (error) {
    console.error('\n❌ Fatal error:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

module.exports = { generateSchemas };
