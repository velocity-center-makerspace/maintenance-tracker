gen:
  @echo "Generating SQL files..."
  sqlc generate

down: (_down "./data/dev.db")

up: (_up "./data/dev.db")

test:
  go test -v ./... | sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''

_down target:
  @echo "Bringing down {{target}}"
  goose sqlite -dir ./db/migrations/ {{target}} down

_up target:
  @echo "Bring down {{target}}"
  goose sqlite -dir ./db/migrations/ {{target}} up
