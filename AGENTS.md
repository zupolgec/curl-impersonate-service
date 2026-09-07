# AGENTS.md

Istruzioni di progetto per agenti AI.

## Documentazione
La documentazione di progetto vive in `/docs`:
- `docs/PLAN.md` — piano di lavoro e stato delle fasi
- `docs/ARCHITECTURE.md` — decisioni architetturali

`/docs` **resta solo in locale** per questo progetto (preferenza confermata da
mt il 2026-07-16). È in `.gitignore` e non va mai committato.

## Convenzioni
- Commit diretti su `main` (niente PR per questo progetto), con tag + release.
- Release: le voci vanno in `## [Unreleased]` di `CHANGELOG.md`, poi
  `scripts/release.sh X.Y.Z` chiude la sezione, committa, tagga e pusha. La
  versione vive solo nel tag: CI la inietta nel binario e nell'immagine, i
  build locali riportano `dev`. Non esiste più una costante da aggiornare.
- Messaggi di commit brevi: una riga di intestazione, no co-authoring.
- Go: mantenere i test eseguibili con `CGO_ENABLED=0` (l'executor CGO è dietro
  build tag, il datastore usa SQLite pure-Go).

## Comandi utili
- Test: `CGO_ENABLED=0 go test ./...`
- Build locale: `go build -o impersonate-service .`
- Build immagine: `docker build -t curl-impersonate-service .`
