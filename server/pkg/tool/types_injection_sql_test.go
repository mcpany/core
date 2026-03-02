// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestCheckSQLiteInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"safe value", "SELECT * FROM users", false},
		{"shell command", ".shell echo 'hacked'", true},
		{"system command", ".system ls -la", true},
		{"open command", ".open /etc/passwd", true},
		{"output command", ".output hack.txt", true},
		{"once command", ".once hack.txt", true},
		{"read command", ".read hack.sql", true},
		{"import command", ".import hack.csv table", true},
		{"load command", ".load mylib.so", true},
		{"upper case shell", ".SHELL echo 'hacked'", true},
		{"space before dot shell", " .shell echo 'hacked'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkSQLiteInjection(tt.val); (err != nil) != tt.wantErr {
				t.Errorf("checkSQLiteInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckMySQLInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"safe value", "SELECT * FROM users", false},
		{"system command", "system ls -la", true},
		{"source command", "source hack.sql", true},
		{"infile access", "SELECT * FROM users INTO OUTFILE '/tmp/hack.txt'", true},
		{"outfile access", "LOAD DATA INFILE '/etc/passwd' INTO TABLE users", true},
		{"mixed case infile", "Load Data InFiLe '/etc/passwd'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkMySQLInjection(tt.val); (err != nil) != tt.wantErr {
				t.Errorf("checkMySQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckPSQLInjection(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		valTrimmed string
		wantErr    bool
	}{
		{"safe value", "SELECT * FROM users", "SELECT * FROM users", false},
		{"shell command", "\\! ls -la", "\\! ls -la", true},
		{"output command", "\\o /tmp/hack.txt", "\\o /tmp/hack.txt", true},
		{"copy command", "\\copy users to '/tmp/hack.txt'", "\\copy users to '/tmp/hack.txt'", true},
		{"copy program", "COPY users TO PROGRAM 'ls -la'", "COPY users TO PROGRAM 'ls -la'", true},
		{"mixed case copy program", "CoPy users To PrOgRaM 'ls -la'", "CoPy users To PrOgRaM 'ls -la'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkPSQLInjection(tt.val, tt.valTrimmed); (err != nil) != tt.wantErr {
				t.Errorf("checkPSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSQLInjection(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		{"safe simple value", "users", "mysql", 1, false},
		{"safe query without keywords boundary", "my_table_update", "psql", 0, false},
		{"keyword without boundary", "my_drop_table", "mysql", 0, false},

		// MySQL specific tests
		{"mysql safe", "SELECT * FROM users", "mysql", 0, true}, // Will trigger checkSQLKeywords
		{"mysql system", "system ls -la", "mysql", 0, true},
		{"mysql infile", "INTO OUTFILE '/tmp/hack'", "mysql", 0, true},

		// SQLite specific tests
		{"sqlite safe", "SELECT *", "sqlite3", 0, true},
		{"sqlite shell", ".shell ls", "sqlite3", 0, true},

		// PSQL specific tests
		{"psql safe", "SELECT *", "psql", 0, true},
		{"psql meta command", "\\! ls -la", "psql", 0, true},

		// Generic SQL Injection keyword tests (from checkSQLKeywords)
		{"keyword OR", "val OR 1=1", "mysql", 0, true},
		{"keyword AND", "val AND 1=1", "mysql", 0, true},
		{"keyword UNION", "val UNION SELECT", "mysql", 0, true},
		{"keyword SELECT", "SELECT *", "mysql", 0, true},
		{"keyword FROM", "FROM users", "mysql", 0, true},
		{"keyword WHERE", "WHERE id=1", "mysql", 0, true},
		{"keyword JOIN", "JOIN users", "mysql", 0, true},
		{"keyword DROP", "DROP TABLE", "mysql", 0, true},
		{"keyword ALTER", "ALTER TABLE", "mysql", 0, true},
		{"keyword CREATE", "CREATE TABLE", "mysql", 0, true},
		{"keyword INSERT", "INSERT INTO", "mysql", 0, true},
		{"keyword UPDATE", "UPDATE users", "mysql", 0, true},
		{"keyword DELETE", "DELETE FROM", "mysql", 0, true},
		{"comment dash", "-- comment", "mysql", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkSQLInjection(tt.val, tt.base, tt.quoteLevel); (err != nil) != tt.wantErr {
				t.Errorf("checkSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
