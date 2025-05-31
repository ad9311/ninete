CREATE TYPE "public"."ledger_status" AS ENUM('n/a', 'pending', 'paid', 'overdue', 'cancelled');--> statement-breakpoint
CREATE TYPE "public"."ledger_type" AS ENUM('budget', 'payable/receivable');--> statement-breakpoint
CREATE TYPE "public"."transaction_category" AS ENUM('housing', 'utilities', 'groceries', 'restaurants', 'foodDelivery', 'transportation', 'healthcare&wellness', 'personalCare', 'shopping', 'entertainment', 'travel&vacations', 'education', 'children&dependents', 'pets', 'gifts&donations', 'financialServices', 'savings&investments', 'workExpenses', 'homeImprovement', 'taxes', 'miscellaneous', 'income', 'payable', 'receivable');--> statement-breakpoint
CREATE TYPE "public"."transaction_type" AS ENUM('credit', 'debit');--> statement-breakpoint
CREATE TABLE "ledgers" (
	"id" serial PRIMARY KEY NOT NULL,
	"user_id" integer NOT NULL,
	"title" text,
	"description" text,
	"year" integer NOT NULL,
	"month" integer NOT NULL,
	"type" "ledger_type" NOT NULL,
	"status" "ledger_status" DEFAULT 'n/a' NOT NULL,
	"total_credits" numeric(10, 2) DEFAULT '0' NOT NULL,
	"total_debits" numeric(10, 2) DEFAULT '0' NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "sessions" (
	"id" text PRIMARY KEY NOT NULL,
	"user_id" integer NOT NULL,
	"expires_at" timestamp with time zone NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "transactions" (
	"id" serial PRIMARY KEY NOT NULL,
	"ledger_id" integer NOT NULL,
	"description" text NOT NULL,
	"amount" numeric(10, 2) NOT NULL,
	"date" timestamp with time zone NOT NULL,
	"category" "transaction_category" NOT NULL,
	"type" "transaction_type" NOT NULL,
	"is_estimated" boolean DEFAULT false NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "users" (
	"id" serial PRIMARY KEY NOT NULL,
	"email" text NOT NULL,
	"username" text NOT NULL,
	"password_hash" text NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT "users_email_unique" UNIQUE("email"),
	CONSTRAINT "users_username_unique" UNIQUE("username")
);
--> statement-breakpoint
ALTER TABLE "ledgers" ADD CONSTRAINT "ledgers_user_id_users_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "sessions" ADD CONSTRAINT "sessions_user_id_users_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "transactions" ADD CONSTRAINT "transactions_ledger_id_ledgers_id_fk" FOREIGN KEY ("ledger_id") REFERENCES "public"."ledgers"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
CREATE UNIQUE INDEX "user_budget_month_year_unique_idx" ON "ledgers" USING btree ("user_id","year","month") WHERE "ledgers"."type" = 'budget';