import { readFile } from 'node:fs/promises';
import path from 'node:path';

import { testdataDir } from './paths';

export async function readCriticalFixture(name: string): Promise<string> {
  return readFile(path.join(testdataDir, name), 'utf8');
}
