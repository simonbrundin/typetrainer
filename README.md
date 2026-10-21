# typetrainer

Ett CLI-baserat skrivträningsverktyg för svenska som anpassar sig efter
dina svaga **bigram** (teckenpar som "sj", "er", "st").

Applikationen mäter fel och tid per teckenpar, rankar de svagaste och
genererar en träningslista som prioriterar dem - spridda över flera
problemområden i stället för att köra samma mönster igen och igen.

## Kom igång

```bash
cd src/go
go build -o ~/.local/bin/typetrainer ./cmd/cli

typetrainer              # adaptiv träning (standard)
typetrainer sprint 30    # 30-sekunderssprint
typetrainer practice sj  # träna ett specifikt teckenpar
typetrainer stats        # statistik
```

## Bygg och test

```bash
cd src/go
go build -o typetrainer ./cmd/cli
go test ./...
```

## Dokumentation

| Dokument | Innehåll |
|---|---|
| [SPEC.md](SPEC.md) | Designen: datamodell, algoritm, konfiguration. |
| [AGENTS.md](AGENTS.md) | Regler för agenter som jobbar i repot. |
| [~/wiki](../agents/wiki/) | Kunskaps-wiki: anteckningar om arkitekturen, kända buggar och beslut. |

## Kunskaps-wiki

Alla repor delar en kunskaps-wiki som ligger i `~/repos/agents/wiki` -
atomiska markdown-anteckningar som är länkade till varandra. Den ger agenter
(och människor) samma orientering utan att läsa hela koden, och den överlever
både kontextfönster och modellbyten.

```bash
wikictl list                      # vad finns det?
wikictl search "bigram"           # sök i anteckningarna
wikictl search "fel" --project typetrainer
wikictl ingest                    # flytta inboxen in i grafen
```

Anteckningar om typetrainer ligger under `notes/` i wikin, taggade med
`projekt:typetrainer`.
