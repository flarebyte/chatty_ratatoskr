import path from 'node:path';

export const repoRoot = path.resolve(import.meta.dir, '../../..');
export const testdataDir = path.join(repoRoot, 'testdata', 'e2e', 'critical');
