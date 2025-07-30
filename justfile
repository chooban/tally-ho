set shell := ["zsh", "-cu"]
set dotenv-required := true
set dotenv-load := true

test-all:
    go test ./...

dldb:
    rm -f ./localdev/db/blog.sqlite

    fly machine start
    flyctl ssh sftp get -a $FLY_APP /data/db/blog.sqlite ./localdev/db/blog.sqlite

updatedb:
    echo "Setting mentions of $PROD_BASE_URL to $BASE_URL"
    sqlite3 ./localdev/db/blog.sqlite "UPDATE mentions SET value = replace(value, 'tally-ho.fly.dev', 'bear-sacred-gannet.ngrok-free.app');"
    sqlite3 ./localdev/db/blog.sqlite "UPDATE entries SET value = replace(value, 'tally-ho.fly.dev', 'bear-sacred-gannet.ngrok-free.app');"
