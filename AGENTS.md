# AGENTS.md

Istruzioni di progetto per agenti AI.

## Documentazione
La documentazione di progetto vive in `/docs`:
- `docs/PLAN.md` — piano di lavoro e stato delle fasi
- `docs/ARCHITECTURE.md` — decisioni architetturali

`/docs` **non** viene committato con il resto del progetto senza chiedere prima
a mt (preferenza non ancora confermata: chiedere ad ogni commit).

## Convenzioni
- Commit diretti su `main` (niente PR per questo progetto), con tag + release.
- Messaggi di commit brevi: una riga di intestazione, no co-authoring.
- Go: mantenere i test eseguibili con `CGO_ENABLED=0` (l'executor CGO è dietro
  build tag, il datastore usa SQLite pure-Go).

## Comandi utili
- Test: `CGO_ENABLED=0 go test ./...`
- Build locale: `go build -o impersonate-service .`
- Build immagine: `docker build -t curl-impersonate-service .`
