import ballerina/io;

type RecordA record {
    int id;
    string name;
};

type RecordB record {
    int id;
    float score;
};

function process(RecordA|RecordB input) returns int {
    if input is RecordA {
        return input.id + 1;
    } else {
        return input.id;
    }
}

public function main() {
    RecordA a = {id: 1, name: "test"};
    RecordB b = {id: 2, score: 3.5};
    io:println(process(a));
    io:println(process(b));
}
