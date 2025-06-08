ALTER TABLE "public"."ledgers" ALTER COLUMN "type" SET DATA TYPE text;--> statement-breakpoint
DROP TYPE "public"."ledger_type";--> statement-breakpoint
CREATE TYPE "public"."ledger_type" AS ENUM('budget', 'payable', 'receivable');--> statement-breakpoint
ALTER TABLE "public"."ledgers" ALTER COLUMN "type" SET DATA TYPE "public"."ledger_type" USING "type"::"public"."ledger_type";