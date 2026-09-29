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

type Outer record {|
    Inner i?;
|};

type OpenOuter record {
    Inner i?;
};

type Inner record {|
    int y?;
    int[] zs?;
|};

public function main() {
    Outer o = {};
    o.i.y = 3;
    io:println(o); // @output {"i":{"y":3}}

    OpenOuter oo = {};
    oo.i.zs[1] = 4;
    io:println(oo); // @output {"i":{"zs":[0,4]}}

    map<Outer> mo = {};
    mo["a"].i.y = 5;
    io:println(mo); // @output {"a":{"i":{"y":5}}}

    Outer[] os = [];
    os[1].i.zs[0] = 6;
    io:println(os); // @output [{},{"i":{"zs":[6]}}]
}
