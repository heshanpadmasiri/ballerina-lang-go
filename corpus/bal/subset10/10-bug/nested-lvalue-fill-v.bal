// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import ballerina/io;

type S record {|
    S[] rs?;
    int x?;
|};

type L record {|
    map<int>[] xs?;
|};

public function main() {
    S s = {};
    s.rs[0]["x"] = 1;
    io:println(s); // @output {"rs":[{"x":1}]}

    map<S[]> ms = {};
    ms["a"][1]["x"] = 1;
    io:println(ms); // @output {"a":[{},{"x":1}]}

    map<map<int>[]> ml = {};
    ml["a"][1]["k"] = 1;
    io:println(ml); // @output {"a":[{},{"k":1}]}

    L r = {};
    r.xs[1]["k"] = 1;
    io:println(r); // @output {"xs":[{},{"k":1}]}

    r.xs[0]["j"] = 2;
    r.xs[1]["k"] = 3;
    io:println(r); // @output {"xs":[{"j":2},{"k":3}]}
}
