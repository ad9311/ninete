package db

import "errors"

var (
	ErrForeignKeysDisabled = errors.New("foreign key enforcement is off on a new connection")
	ErrPragmaNoValue       = errors.New("pragma returned no value")
	ErrUnknownEnvStamp     = errors.New("no environment stamp defined for ENV")
	ErrEnvStampRead        = errors.New("failed to read the database environment stamp")
	ErrEnvStampWrite       = errors.New("failed to write the database environment stamp")
	ErrEnvStampMismatch    = errors.New("database environment stamp does not match ENV")
	ErrSnapshotFailed      = errors.New("failed to write database snapshot")
	ErrSnapshotDir         = errors.New("failed to prepare the snapshot directory")
	ErrSnapshotPrune       = errors.New("failed to prune old snapshots")
	ErrSnapshotNameTaken   = errors.New("could not find a free snapshot filename")
	ErrDBVersionRead       = errors.New("failed to read the database migration version")
	ErrSchemaVersionRead   = errors.New("failed to read the embedded migration versions")
)
