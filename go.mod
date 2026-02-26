module cryptobar

go 1.23

require (
	github.com/caseymrm/menuet v1.0.3
	github.com/gorilla/websocket v1.5.3
)

require github.com/caseymrm/askm v1.0.0 // indirect

replace github.com/caseymrm/menuet => ./internal/menuet
