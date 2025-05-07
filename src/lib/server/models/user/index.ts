import { createInsertSchema, createSelectSchema } from 'drizzle-zod';
import { db } from '$lib/server/db';
import { usersTable, type User } from '$lib/server/db/schema';
import { z } from 'zod';
import { eq } from 'drizzle-orm';
import { hash, verify } from '@node-rs/argon2';

const USERNAME_VALIDATION_MESSAGES = {
	min: 'Username must be at least 3 characters long',
	max: 'Username cannot exceed 50 characters'
} as const;

const EMAIL_VALIDATION_MESSAGE = 'Invalid email address';

const PASSWORD_VALIDATION_MESSAGES = {
	min: 'Password must be at least 8 characters long',
	max: 'Password must be less than 20 characters long',
	nonempty: 'Password is required'
} as const;

export const userCreateSchema = createInsertSchema(usersTable, {
	username: (schema) =>
		schema
			.min(3, { message: USERNAME_VALIDATION_MESSAGES.min })
			.max(20, { message: USERNAME_VALIDATION_MESSAGES.max }),
	email: (schema) => schema.email({ message: EMAIL_VALIDATION_MESSAGE }),
	passwordHash: (schema) => schema.nonempty({ message: PASSWORD_VALIDATION_MESSAGES.nonempty })
});

export const userSelectSchema = createSelectSchema(usersTable, {
	id: (schema) => schema,
	username: (schema) => schema,
	email: (schema) => schema
}).omit({ passwordHash: true, createdAt: true, updatedAt: true });

export const userRegistraterSchema = userCreateSchema
	.extend(userCreateSchema.shape)
	.omit({
		id: true,
		passwordHash: true,
		createdAt: true,
		updatedAt: true
	})
	.extend({
		password: z
			.string()
			.min(8, { message: PASSWORD_VALIDATION_MESSAGES.min })
			.max(20, { message: PASSWORD_VALIDATION_MESSAGES.max })
	});

export type UserCreateData = z.infer<typeof userCreateSchema>;
export type UserRegistrationData = z.infer<typeof userRegistraterSchema>;
export type UserSelectData = z.infer<typeof userSelectSchema>;

export async function createUser(data: UserRegistrationData) {
	userRegistraterSchema.parse(data);

	const passwordHash = await hash(data.password, {
		memoryCost: 19456,
		timeCost: 2,
		outputLen: 32,
		parallelism: 1
	});

	const newUser: UserCreateData = {
		username: data.username,
		email: data.email,
		passwordHash
	};

	userCreateSchema.parse(newUser);

	const result = await db.insert(usersTable).values(newUser).returning();

	return result[0];
}

export async function findUserById(id: number): Promise<UserSelectData | undefined> {
	const rows = await db
		.select({ id: usersTable.id, username: usersTable.username, email: usersTable.email })
		.from(usersTable)
		.where(eq(usersTable.id, id));

	const user = rows[0];
	userSelectSchema.parse(user);

	return user;
}

export async function findUserByEmail(email: string): Promise<User | undefined> {
	const rows = await db.select().from(usersTable).where(eq(usersTable.email, email));

	const user = rows[0];

	return user;
}

export async function verfifyPassword(passwordHash: string, password: string): Promise<boolean> {
	return await verify(passwordHash, password, {
		memoryCost: 19456,
		timeCost: 2,
		outputLen: 32,
		parallelism: 1
	});
}
