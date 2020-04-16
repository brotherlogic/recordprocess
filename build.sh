protoc --proto_path ../../../ -I=./proto --go_out=plugins=grpc:./proto proto/recordprocess.proto
mv proto/github.com/brotherlogic/recordprocess/proto/* ./proto
