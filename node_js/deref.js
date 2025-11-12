import fs from 'fs/promises';
import { dereference } from '@scalar/openapi-parser';
import yaml from 'js-yaml';
import stringify from 'fast-safe-stringify';

const args = process.argv.slice(2);
if (args.length === 0) process.exit(1);

const specPath = args[0];

try {
    const raw = await fs.readFile(specPath, 'utf-8');

    let specification;
    try {
        specification = JSON.parse(raw);
    } catch {
        specification = yaml.load(raw);
    }

    const { schema } = await dereference(specification);

    console.log(stringify(schema, null, 2)); // circular-safe
} catch (err) {
    console.error(err);
    process.exit(1);
}
