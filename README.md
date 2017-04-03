# reicheltApi
Ein simpler reichelt.de-client (geschrieben in go)

## Was er tut
Unter zuhilfenahme der reichelt.de-autocompletion der Suche wird eine Reichelt-Suche implementiert.
Die Funktion FindPart() nimmt einen Suchquery entgegen und liefert eine Teileliste zurück.

Mit den zurückgelieferten Teilenummern kann man:
- Das Produktbild abfragen (Connection.GetImage)
- Technische Daten fetchen (Connection.GetMeta)
- Den aktuellen Preis des Teils abfragen (Connection.GetPrice)

### Connection.GetMeta()
folgende Metadaten werden immer ergänzt:
- datasheets: Eine Liste alle Datenblätter
- Manufaturer:
  - Name: Der Name des Herstellers
  - Partnumber: Die Manufacturer Part Number (MPN), mit der man das Teil in anderen Systemen findet (z.B. auf [octopart](octopart.com))


## ApiServer
in cmd/ApiServer liegt eine Beispielanwendung, welche mit Hilfe der Library eine kleine API Hostet:

`[addr]/search/[query]`: Führt den Query aus, liefert die Ergebnisse in JSON zurück
`[addr]/meta/[id]`: Liefert alle Metadaten des Parts mit der `id` zurück.
`[addr]/image/[id]`: Liefert das Bild für das Part mit der `id` zurück.
`[addr]/price/[id]`: Liefert den aktuellen Preis in Euro zurück.

Der ApiServer cached alle Daten bis auf den Preis für immer, dass kann mit 
flags geändert werden.
