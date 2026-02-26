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

const AND = 12 & 10;
const OR = 12 | 10;
const XOR = 12 ^ 10;
const SHL = 1 << 3;
const SHR = 16 >> 2;

public function main() {
    io:println(AND); // @output 8
    io:println(OR); // @output 14
    io:println(XOR); // @output 6
    io:println(SHL); // @output 8
    io:println(SHR); // @output 4
}
