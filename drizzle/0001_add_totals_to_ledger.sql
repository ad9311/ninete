ALTER TABLE "ledgers" ADD COLUMN "total_credits" numeric(10, 2) DEFAULT '0' NOT NULL;--> statement-breakpoint
ALTER TABLE "ledgers" ADD COLUMN "total_debits" numeric(10, 2) DEFAULT '0' NOT NULL;