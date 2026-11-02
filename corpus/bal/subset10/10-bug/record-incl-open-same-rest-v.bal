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

type OpenA record {
    int x;
};

type OpenB record {
    string y;
};

type ExplicitA record {|
    int a;
    int...;
|};

type ExplicitB record {|
    int b;
    int...;
|};

type Both record {|
    *OpenA;
    *OpenB;
|};

type BothExplicit record {|
    *ExplicitA;
    *ExplicitB;
|};

public function main() {
    Both both = {x: 1, y: "hello", "z": 2.5};
    io:println(both); // @output {"x":1,"y":"hello","z":2.5}

    BothExplicit explicit = {a: 1, b: 2, "c": 3};
    io:println(explicit); // @output {"a":1,"b":2,"c":3}
}
