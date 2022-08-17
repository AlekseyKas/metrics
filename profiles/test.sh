while true
do
  curl 127.0.0.1:8080
done
# go test -bench=. -memprofile=profiles/base.pprof
# go tool pprof -http=":8090" profiles/base.pprof