import { drizzle } from 'drizzle-orm/node-postgres';
import { env } from '$env/dynamic/private';
import { Pool } from 'pg';
import * as schema from './schema';

if (!env.DATABASE_URL) throw new Error('DB INDEX: DATABASE_URL is not set');

const pool = new Pool({
	connectionString: env.DATABASE_URL
});

export const db = drizzle({ client: pool, schema });
export type DBTransaction = Parameters<Parameters<typeof db.transaction>[0]>[0];
