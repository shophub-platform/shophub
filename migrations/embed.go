// Package migrations ugrađuje SQL migracije u binarni fajl kako bi ih
// golang-migrate pokretao bez potrebe za fajl-sistemom u produkciji.
package migrations

import "embed"

// FS sadrži sve *.sql migracije (par up/down po verziji).
//
//go:embed *.sql
var FS embed.FS
