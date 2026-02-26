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

const A = 10 + 5;
const B = 20 - 3;
const C = 4 * 6;
const D = 100 / 10;
const E = 17 % 5;

public function main() {
    io:println(A); // @output 15
    io:println(B); // @output 17
    io:println(C); // @output 24
    io:println(D); // @output 10
    io:println(E); // @output 2
}
