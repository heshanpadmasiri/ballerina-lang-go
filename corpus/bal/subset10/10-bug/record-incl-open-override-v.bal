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

type IntRest record {|
    int a;
    int...;
|};

type StringRest record {|
    string b;
    string...;
|};

type OpenOverride record {
    *IntRest;
    *StringRest;
};

type Derived record {|
    *OpenOverride;
|};

public function main() {
    Derived d = {a: 1, b: "s", "z": [1, 2]};
    io:println(d); // @output {"a":1,"b":"s","z":[1,2]}
}
