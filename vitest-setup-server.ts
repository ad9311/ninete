/* eslint-disable drizzle/enforce-delete-with-where */

import { vi } from 'vitest';
import { beforeAll, afterAll, beforeEach } from 'vitest';
import postgres from 'postgres';
import { drizzle, PostgresJsDatabase } from 'drizzle-orm/postgres-js';
import * as shelljs from 'shelljs';
import * as schema from './src/lib/server/db/schema';

const databaseUrl = process.env.DATABASE_URL_TEST;
if (!databaseUrl) {
	throw new Error('DATABASE_URL_TEST environment variable is not set');
}

// Initialize the postgres client and Drizzle instance for the test database
const testClient = postgres(databaseUrl);
export const testDb: PostgresJsDatabase<typeof schema> = drizzle(testClient, { schema });

// Mock the application's primary db module ($lib/server/db)
// to use our testDb instance whenever it's imported during tests.
vi.mock('$lib/server/db', () => ({
	db: testDb
	// If $lib/server/db exports other things, ensure they are handled here if needed
	// For example, by spreading ...await importActual('$lib/server/db') if it had other named exports.
	// Since it only exports 'db', this is sufficient.
}));

beforeAll(async () => {
	console.log('Verifying test database connection...');
	try {
		await testClient`SELECT 1`;
		console.log('Test database connection successful.');
	} catch (error) {
		console.error('Failed to connect to test database:', error);
		throw error;
	}

	console.log('Pushing schema to test database...');
	const command = `npx drizzle-kit push \
        --dialect=postgresql \
        --schema=./src/lib/server/db/schema.ts \
        --url='${databaseUrl}'`;

	const result = shelljs.exec(command, { silent: true });

	if (result.code !== 0) {
		console.error('Schema push failed:');
		console.error('stderr:', result.stderr);
		console.error('stdout:', result.stdout);
		throw new Error('Failed to push schema to database');
	}
	console.log('Schema push completed successfully.');
}, 60000); // Increased timeout for setup

// Clear relevant tables before each test to ensure a clean state
beforeEach(async () => {
	await testDb.delete(schema.sessionsTable).catch(() => {});
	await testDb.delete(schema.usersTable).catch(() => {});
	// console.log('Test database tables cleared.');
});

afterAll(async () => {
	if (testClient) {
		console.log('Closing test database connection...');
		await testClient.end();
		console.log('Test database connection closed.');
	}
}, 30000); // Increased timeout for teardown
