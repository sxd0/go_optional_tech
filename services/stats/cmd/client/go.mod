module go_optional_tech/services/stats/cmd/client

go 1.22

require (
    google.golang.org/grpc v1.64.0
    go_optional_tech/proto/statspb v0.0.0
)

replace go_optional_tech/proto/statspb => ../../../../proto/statspb
