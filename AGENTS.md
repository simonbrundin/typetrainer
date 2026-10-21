# AGENTS.md — typetrainer

Regler för agenter som jobbar i det här repot.

> **Globala regler:** det finns också en global `AGENTS.md` (`~/AGENTS.md`)
> som gäller i alla repor. Den innehåller de två wiki-reglerna och andra
> gemensamma regler. Den här filen gäller **specifikt typetrainer**.

## Wiki: läs först, skriv sedan

Det finns en gemensam kunskaps-wiki som används av alla repor. Den ligger i
`~/wiki` och **versioneras i `~/repos/agents/wiki`**, så att kunskapen
överlever tillsammans med koden - och följer med när kunskapen sprids till
nya repor.

Wikin innehåller anteckningar om arkitekturen, algoritmerna, kända buggar
och de beslut som fattats - information som annars bara finns i någons
huvud eller i en gammal chatt.

### 1. Läs wikin innan du frågar användaren

> Innan du frågar användaren, leta i wikin först - det finns en bra chans
> att svaret redan finns där.

```bash
wikictl list                       # vad finns det?
wikictl search "bigram"            # sök i titlar, taggar och brödtext
wikictl search "fel" --project typetrainer
wikictl show notes/arkitektur      # läs en specifik anteckning
```

(`wikictl` är installerat globalt och hittar wikin från valfri katalog.)

Läs `~/wiki/index.md` för en överblick, `~/wiki/catalog.md` för alla sidor och
`~/wiki/graph.md` för hur anteckningar hänger ihop. **Läs hela anteckningen**, inte bara träffraden - noterna
beskriver ofta *varför*, och det är den delen som sparar tid.

Att fråga användaren om något som redan står i wikin är en av de dyraste
felen en agent kan göra. Läs först.

### 2. Skriv en not när något stort händer

> Varje gång något stort händer - varje incident, varje nytt projekt, varje
> stort beslut - lämna en kort not i wikin inbox.

Det gäller även om inget gick fel. Ett beslut, en ny feature, en
icke-obvious bugg, en ändring i arbetssättet - allt som en framtida agent
skulle ha svårt att härleda ur koden.

```bash
wikictl note \
  --title "Rubrik på noten" \
  --tags "bug, session" \
  --body "Vad som hände, varför, och vad som gäller framöver."

wikictl ingest --dry-run   # se vad som skulle hända
wikictl ingest              # gör det
```

Noter taggas automatiskt med `projekt:typetrainer`.

Regler för noterna:

- **En idé per not.** Två saker = två noter.
- **Länka hellre än att duplicera.** Om något redan finns, länka till den
  anteckningen.
- **Skriv varför, inte bara vad.** Koden visar vad - noten ska visa
  varför.
- **Skriv på svenska**, i linje med resten av repot.
- **Inga hemligheter.** Noter versionshålls i git - inga credentials,
  tokens eller personuppgifter.

### 3. Sök i wikin när du är osäker

Om du är osäker på hur något fungerar, om ett beslut, eller om varför kod
ser ut som den gör - sök i wikin innan du gissar. Saknar svaret? Då ska
det in i wikin, se punkt 2.

## Projektfakta

TypeTrainer är en CLI-app (cobra) med bubbletea-TUI som tränar svensk
skrivning genom att prioritera svaga bigram (teckenpar).

```bash
cd src/go
go build -o typetrainer ./cmd/cli   # bygg (INTE "go build .")
go test ./...                        # kör testerna

All Go-kod ligger i `src/go` (modul `typetrainer`). Entrypointen är
`src/go/cmd/cli/main.go` - det finns ingen `main.go` i modulroten.

## Konventioner

- **Svenska** i UI-strängar, kommentarer och commit-meddelanden.
  **Engelska** för identifierare och typnamn.
- Varje paket i `internal/` har ett ansvar: `bigram` (ren domänstatistik),
  `wordlist` (ord och urval), `stats` (persistens och analys), `session`
  (TUI-orchestrering). `session` är det enda paketet som känner till
  bubbletea.
- Tester ligger bredvid koden, i samma paket, och är rena enhetstester utan
  mocks.
- **Ändra bara det som behövs.** Rör inte angränsande kod, kommentarer
  eller formatering som inte ingår i uppgiften.

## Vad du inte ska göra

- Ändra `~/wiki/catalog.md` eller `~/wiki/graph.md` för hand - de är
  genererade av `wikictl index`. `~/wiki/index.md` är kuraterad och
  sköts för hand.
- Ta bort en anteckning för att den känns gammal. Markera den som
  `status: deprecated` och skriv varför.
- Skriva hemligheter i wikin, i commit-meddelanden eller i kod.
- Lägga till beroenden utan att det behövs. repot har tre:
  cobra, bubbletea, lipgloss.

## Mer

- `~/wiki/index.md` - den kuraterade katalogen
- `~/wiki/catalog.md` - alla sidor i wikin (genererad)
- `~/wiki/graph.md` - kunskapsgrafen (genererad)
- `~/wiki/AGENTS.md` - schemat: sidtyper, frontmatter, arbetsflöden
- `SPEC.md` - designdokumentet (data modell, algoritm, konfiguration).
  **Delvis inaktuellt** - koden har den sista ordningen.

## Andra repon

Andra repon delar samma wiki. För att ge ett annat repo wikinreglerna:

```bash
wikictl init /path/to/repo
```

Det skriver in wikinblocket i repots `AGENTS.md` (idempotent - det ersätter
ett tidigare block om texten ändrats). Anteckningar i det repot taggas då
automatiskt med `projekt:<repo>`.
