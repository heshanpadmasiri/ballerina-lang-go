package common

import (
	"ballerina-lang-go/tools/diagnostics"
)

// FIXME: make this private
type ParserRuleContext struct {
	value string
}

func (p ParserRuleContext) String() string {
	return p.value
}

var (
	// Productions
	PARSER_RULE_CONTEXT_COMP_UNIT                                  = ParserRuleContext{value: "comp-unit"}
	PARSER_RULE_CONTEXT_EOF                                        = ParserRuleContext{value: "eof"}
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE                             = ParserRuleContext{value: "top-level-node"}
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_METADATA            = ParserRuleContext{value: "top-level-node-without-metadata"}
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_MODIFIER            = ParserRuleContext{value: "top-level-node-without-modifier"}
	PARSER_RULE_CONTEXT_FUNC_DEF                                   = ParserRuleContext{value: "func-def"}
	PARSER_RULE_CONTEXT_FUNC_DEF_START                             = ParserRuleContext{value: "function-def-start"}
	PARSER_RULE_CONTEXT_FUNC_DEF_OR_FUNC_TYPE                      = ParserRuleContext{value: "func-def-or-func-type"}
	PARSER_RULE_CONTEXT_FUNC_DEF_FIRST_QUALIFIER                   = ParserRuleContext{value: "func-def-first-qualifier"}
	PARSER_RULE_CONTEXT_FUNC_DEF_SECOND_QUALIFIER                  = ParserRuleContext{value: "func-def-second-qualifier"}
	PARSER_RULE_CONTEXT_FUNC_DEF_WITHOUT_FIRST_QUALIFIER           = ParserRuleContext{value: "func-def-without-first-qualifier"}
	PARSER_RULE_CONTEXT_PARAM_LIST                                 = ParserRuleContext{value: "parameters"}
	PARSER_RULE_CONTEXT_PARAMETER_START                            = ParserRuleContext{value: "parameter-start"}
	PARSER_RULE_CONTEXT_PARAMETER_START_WITHOUT_ANNOTATION         = ParserRuleContext{value: "parameter-start-without-annotation"}
	PARSER_RULE_CONTEXT_PARAM_END                                  = ParserRuleContext{value: "param-end"}
	PARSER_RULE_CONTEXT_REQUIRED_PARAM                             = ParserRuleContext{value: "required-parameter"}
	PARSER_RULE_CONTEXT_DEFAULTABLE_PARAM                          = ParserRuleContext{value: "defaultable-parameter"}
	PARSER_RULE_CONTEXT_REST_PARAM                                 = ParserRuleContext{value: "rest-parameter"}
	PARSER_RULE_CONTEXT_PARAM_START                                = ParserRuleContext{value: "parameter-start"}
	PARSER_RULE_CONTEXT_PARAM_RHS                                  = ParserRuleContext{value: "param-rhs"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_PARAM_RHS                        = ParserRuleContext{value: "function-type-desc-param-rhs"}
	PARSER_RULE_CONTEXT_REST_PARAM_RHS                             = ParserRuleContext{value: "rest-param-rhs"}
	PARSER_RULE_CONTEXT_AFTER_PARAMETER_TYPE                       = ParserRuleContext{value: "after-parameter-type"}
	PARSER_RULE_CONTEXT_PARAMETER_NAME_RHS                         = ParserRuleContext{value: "parameter-name-rhs"}
	PARSER_RULE_CONTEXT_REQUIRED_PARAM_NAME_RHS                    = ParserRuleContext{value: "required-param-name-rhs"}
	PARSER_RULE_CONTEXT_FUNC_OPTIONAL_RETURNS                      = ParserRuleContext{value: "func-optional-returns"}
	PARSER_RULE_CONTEXT_FUNC_BODY                                  = ParserRuleContext{value: "func-body"}
	PARSER_RULE_CONTEXT_FUNC_BODY_OR_TYPE_DESC_RHS                 = ParserRuleContext{value: "func-body-or-type-desc-rhs"}
	PARSER_RULE_CONTEXT_ANON_FUNC_BODY                             = ParserRuleContext{value: "annon-func-body"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_END                         = ParserRuleContext{value: "func-type-desc-end"}
	PARSER_RULE_CONTEXT_EXTERNAL_FUNC_BODY                         = ParserRuleContext{value: "external-func-body"}
	PARSER_RULE_CONTEXT_EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS         = ParserRuleContext{value: "external-func-body-optional-annots"}
	PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK                            = ParserRuleContext{value: "func-body-block"}
	PARSER_RULE_CONTEXT_MODULE_TYPE_DEFINITION                     = ParserRuleContext{value: "type-definition"}
	PARSER_RULE_CONTEXT_MODULE_CLASS_DEFINITION                    = ParserRuleContext{value: "class-definition"}
	PARSER_RULE_CONTEXT_MODULE_CLASS_DEFINITION_START              = ParserRuleContext{value: "class-definition-start"}
	PARSER_RULE_CONTEXT_FIRST_CLASS_TYPE_QUALIFIER                 = ParserRuleContext{value: "first-class-type-qualifier"}
	PARSER_RULE_CONTEXT_SECOND_CLASS_TYPE_QUALIFIER                = ParserRuleContext{value: "second-class-type-qualifier"}
	PARSER_RULE_CONTEXT_THIRD_CLASS_TYPE_QUALIFIER                 = ParserRuleContext{value: "third-class-type-qualifier"}
	PARSER_RULE_CONTEXT_FOURTH_CLASS_TYPE_QUALIFIER                = ParserRuleContext{value: "fourth-class-type-qualifier"}
	PARSER_RULE_CONTEXT_CLASS_DEF_WITHOUT_FIRST_QUALIFIER          = ParserRuleContext{value: "class-def-without-first-qualifier"}
	PARSER_RULE_CONTEXT_CLASS_DEF_WITHOUT_SECOND_QUALIFIER         = ParserRuleContext{value: "class-def-without-second-qualifier"}
	PARSER_RULE_CONTEXT_CLASS_DEF_WITHOUT_THIRD_QUALIFIER          = ParserRuleContext{value: "class-def-without-third-qualifier"}
	PARSER_RULE_CONTEXT_FIELD_OR_REST_DESCIPTOR_RHS                = ParserRuleContext{value: "field-or-rest-descriptor-rhs"}
	PARSER_RULE_CONTEXT_FIELD_DESCRIPTOR_RHS                       = ParserRuleContext{value: "field-descriptor-rhs"}
	PARSER_RULE_CONTEXT_RECORD_BODY_START                          = ParserRuleContext{value: "record-body-start"}
	PARSER_RULE_CONTEXT_RECORD_BODY_END                            = ParserRuleContext{value: "record-body-end"}
	PARSER_RULE_CONTEXT_RECORD_FIELD                               = ParserRuleContext{value: "record-field"}
	PARSER_RULE_CONTEXT_RECORD_FIELD_OR_RECORD_END                 = ParserRuleContext{value: "record-field-orrecord-end"}
	PARSER_RULE_CONTEXT_RECORD_FIELD_START                         = ParserRuleContext{value: "record-field-start"}
	PARSER_RULE_CONTEXT_RECORD_FIELD_WITHOUT_METADATA              = ParserRuleContext{value: "record-field-without-metadata"}
	PARSER_RULE_CONTEXT_TYPE_DESCRIPTOR                            = ParserRuleContext{value: "type-descriptor"}
	PARSER_RULE_CONTEXT_TYPE_DESC_WITHOUT_ISOLATED                 = ParserRuleContext{value: "type-desc-without-isolated"}
	PARSER_RULE_CONTEXT_CLASS_DESCRIPTOR                           = ParserRuleContext{value: "class-descriptor"}
	PARSER_RULE_CONTEXT_RECORD_TYPE_DESCRIPTOR                     = ParserRuleContext{value: "record-type-desc"}
	PARSER_RULE_CONTEXT_TYPE_REFERENCE                             = ParserRuleContext{value: "type-reference"}
	PARSER_RULE_CONTEXT_TYPE_REFERENCE_IN_TYPE_INCLUSION           = ParserRuleContext{value: "type-reference-in-type-inclusion"}
	PARSER_RULE_CONTEXT_SIMPLE_TYPE_DESC_IDENTIFIER                = ParserRuleContext{value: "simple-type-desc-identifier"}
	PARSER_RULE_CONTEXT_ARG_LIST_OPEN_PAREN                        = ParserRuleContext{value: "("}
	PARSER_RULE_CONTEXT_ARG_LIST                                   = ParserRuleContext{value: "arguments"}
	PARSER_RULE_CONTEXT_ARG_START                                  = ParserRuleContext{value: "argument-start"}
	PARSER_RULE_CONTEXT_ARG_END                                    = ParserRuleContext{value: "arg-end"}
	PARSER_RULE_CONTEXT_ARG_LIST_END                               = ParserRuleContext{value: "argument-end"}
	PARSER_RULE_CONTEXT_ARG_LIST_CLOSE_PAREN                       = ParserRuleContext{value: ")"}
	PARSER_RULE_CONTEXT_ARG_START_OR_ARG_LIST_END                  = ParserRuleContext{value: "arg-start-or-args-list-end"}
	PARSER_RULE_CONTEXT_NAMED_OR_POSITIONAL_ARG_RHS                = ParserRuleContext{value: "named-or-positional-arg"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_DESCRIPTOR                     = ParserRuleContext{value: "object-type-desc"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_MEMBER                  = ParserRuleContext{value: "object-constructor-member"}
	PARSER_RULE_CONTEXT_CLASS_MEMBER                               = ParserRuleContext{value: "class-member"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_MEMBER                         = ParserRuleContext{value: "object-type-member"}
	PARSER_RULE_CONTEXT_CLASS_MEMBER_OR_OBJECT_MEMBER_START        = ParserRuleContext{value: "class-member-or-object-member-start"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_MEMBER_START            = ParserRuleContext{value: "object-constructor-member-start"}
	PARSER_RULE_CONTEXT_CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META = ParserRuleContext{value: "class-member-or-object-member-without-metadata"}
	PARSER_RULE_CONTEXT_OBJECT_CONS_MEMBER_WITHOUT_META            = ParserRuleContext{value: "object-constructor-member-without-metadata"}
	PARSER_RULE_CONTEXT_OBJECT_FUNC_OR_FIELD                       = ParserRuleContext{value: "object-func-or-field"}
	PARSER_RULE_CONTEXT_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY    = ParserRuleContext{value: "object-func-or-field-without-visibility"}
	PARSER_RULE_CONTEXT_OBJECT_MEMBER_VISIBILITY_QUAL              = ParserRuleContext{value: "object-member-visibility-qual"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_START                        = ParserRuleContext{value: "object-method-start"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_FIRST_QUALIFIER              = ParserRuleContext{value: "object-method-first-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_SECOND_QUALIFIER             = ParserRuleContext{value: "object-method-second-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_THIRD_QUALIFIER              = ParserRuleContext{value: "object-method.third-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_FOURTH_QUALIFIER             = ParserRuleContext{value: "object-method-fourth-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER      = ParserRuleContext{value: "object.method.without.first.qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER     = ParserRuleContext{value: "object.method.without.transactional"}
	PARSER_RULE_CONTEXT_OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER      = ParserRuleContext{value: "object.method.without.isolated"}
	PARSER_RULE_CONTEXT_OBJECT_FIELD_START                         = ParserRuleContext{value: "object-field-start"}
	PARSER_RULE_CONTEXT_OBJECT_FIELD_QUALIFIER                     = ParserRuleContext{value: "object-field-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_FIELD_RHS                           = ParserRuleContext{value: "object-field-rhs"}
	PARSER_RULE_CONTEXT_OPTIONAL_FIELD_INITIALIZER                 = ParserRuleContext{value: "optional-field-initializer"}
	PARSER_RULE_CONTEXT_ON_FAIL_OPTIONAL_BINDING_PATTERN           = ParserRuleContext{value: "on-fail-optional-binding-pattern"}
	PARSER_RULE_CONTEXT_FIRST_OBJECT_TYPE_QUALIFIER                = ParserRuleContext{value: "first-object-type-qualifier"}
	PARSER_RULE_CONTEXT_SECOND_OBJECT_TYPE_QUALIFIER               = ParserRuleContext{value: "second-object-type-qualifier"}
	PARSER_RULE_CONTEXT_FIRST_OBJECT_CONS_QUALIFIER                = ParserRuleContext{value: "first-object-cons-qualifier"}
	PARSER_RULE_CONTEXT_SECOND_OBJECT_CONS_QUALIFIER               = ParserRuleContext{value: "second-object-cons-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_CONS_WITHOUT_FIRST_QUALIFIER        = ParserRuleContext{value: "object-cons-without-first-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER        = ParserRuleContext{value: "object-type-without-first-qualifier"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_START                          = ParserRuleContext{value: "object-type-start"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_START                   = ParserRuleContext{value: "object-constructor-start"}
	PARSER_RULE_CONTEXT_IMPORT_DECL                                = ParserRuleContext{value: "import-decl"}
	PARSER_RULE_CONTEXT_IMPORT_ORG_OR_MODULE_NAME                  = ParserRuleContext{value: "import-org-or-module-name"}
	PARSER_RULE_CONTEXT_IMPORT_MODULE_NAME                         = ParserRuleContext{value: "module-name"}
	PARSER_RULE_CONTEXT_IMPORT_PREFIX                              = ParserRuleContext{value: "import-prefix"}
	PARSER_RULE_CONTEXT_IMPORT_PREFIX_DECL                         = ParserRuleContext{value: "import-alias"}
	PARSER_RULE_CONTEXT_IMPORT_DECL_ORG_OR_MODULE_NAME_RHS         = ParserRuleContext{value: "import-decl-org-or-module-name-rhs"}
	PARSER_RULE_CONTEXT_AFTER_IMPORT_MODULE_NAME                   = ParserRuleContext{value: "after-import-module-name"}
	PARSER_RULE_CONTEXT_SERVICE_DECL                               = ParserRuleContext{value: "service-decl"}
	PARSER_RULE_CONTEXT_SERVICE_DECL_START                         = ParserRuleContext{value: "service-decl-start"}
	PARSER_RULE_CONTEXT_SERVICE_DECL_QUALIFIER                     = ParserRuleContext{value: "service-decl-qualifier"}
	PARSER_RULE_CONTEXT_SERVICE_DECL_OR_VAR_DECL                   = ParserRuleContext{value: "service-decl-or-var-decl"}
	PARSER_RULE_CONTEXT_SERVICE_VAR_DECL_RHS                       = ParserRuleContext{value: "service-var-decl-rhs"}
	PARSER_RULE_CONTEXT_OPTIONAL_SERVICE_DECL_TYPE                 = ParserRuleContext{value: "optional-service-decl-type"}
	PARSER_RULE_CONTEXT_OPTIONAL_ABSOLUTE_PATH                     = ParserRuleContext{value: "optional-absolute-path"}
	PARSER_RULE_CONTEXT_ABSOLUTE_RESOURCE_PATH                     = ParserRuleContext{value: "absolute-resource-path"}
	PARSER_RULE_CONTEXT_ABSOLUTE_RESOURCE_PATH_START               = ParserRuleContext{value: "absolute-resource-path-start"}
	PARSER_RULE_CONTEXT_ABSOLUTE_PATH_SINGLE_SLASH                 = ParserRuleContext{value: "absolute-path-single-slash"}
	PARSER_RULE_CONTEXT_ABSOLUTE_RESOURCE_PATH_END                 = ParserRuleContext{value: "absolute-resource-path-end"}
	PARSER_RULE_CONTEXT_SERVICE_DECL_RHS                           = ParserRuleContext{value: "service-decl-rhs"}
	PARSER_RULE_CONTEXT_LISTENERS_LIST                             = ParserRuleContext{value: "listeners-list"}
	PARSER_RULE_CONTEXT_LISTENERS_LIST_END                         = ParserRuleContext{value: "listeners-list-end"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_BLOCK                   = ParserRuleContext{value: "object-constructor-block"}
	PARSER_RULE_CONTEXT_RESOURCE_KEYWORD_RHS                       = ParserRuleContext{value: "resource-keyword-rhs"}
	PARSER_RULE_CONTEXT_OPTIONAL_RELATIVE_PATH                     = ParserRuleContext{value: "optional-relative-path"}
	PARSER_RULE_CONTEXT_RELATIVE_RESOURCE_PATH                     = ParserRuleContext{value: "relative-resource-path"}
	PARSER_RULE_CONTEXT_RELATIVE_RESOURCE_PATH_START               = ParserRuleContext{value: "relative-resource-path-start"}
	PARSER_RULE_CONTEXT_RESOURCE_PATH_SEGMENT                      = ParserRuleContext{value: "resource-path-segment"}
	PARSER_RULE_CONTEXT_RESOURCE_PATH_PARAM                        = ParserRuleContext{value: "resource-path-param"}
	PARSER_RULE_CONTEXT_PATH_PARAM_OPTIONAL_ANNOTS                 = ParserRuleContext{value: "path-param-optional-annots"}
	PARSER_RULE_CONTEXT_PATH_PARAM_ELLIPSIS                        = ParserRuleContext{value: "path-param-ellipsis"}
	PARSER_RULE_CONTEXT_OPTIONAL_PATH_PARAM_NAME                   = ParserRuleContext{value: "optional-path-param-name"}
	PARSER_RULE_CONTEXT_RELATIVE_RESOURCE_PATH_END                 = ParserRuleContext{value: "relative-resource-path-end"}
	PARSER_RULE_CONTEXT_RESOURCE_PATH_END                          = ParserRuleContext{value: "relative-resource-path-end"}
	PARSER_RULE_CONTEXT_RESOURCE_ACCESSOR_DEF_OR_DECL_RHS          = ParserRuleContext{value: "resource-accessor-def-or-decl-rhs"}
	PARSER_RULE_CONTEXT_LISTENER_DECL                              = ParserRuleContext{value: "listener-decl"}
	PARSER_RULE_CONTEXT_CONSTANT_DECL                              = ParserRuleContext{value: "const-decl"}
	PARSER_RULE_CONTEXT_CONST_DECL_TYPE                            = ParserRuleContext{value: "const-decl-type"}
	PARSER_RULE_CONTEXT_CONST_DECL_RHS                             = ParserRuleContext{value: "const-decl-rhs"}
	PARSER_RULE_CONTEXT_NIL_TYPE_DESCRIPTOR                        = ParserRuleContext{value: "nil-type-descriptor"}
	PARSER_RULE_CONTEXT_OPTIONAL_TYPE_DESCRIPTOR                   = ParserRuleContext{value: "optional-type-descriptor"}
	PARSER_RULE_CONTEXT_ARRAY_TYPE_DESCRIPTOR                      = ParserRuleContext{value: "array-type-descriptor"}
	PARSER_RULE_CONTEXT_ARRAY_LENGTH                               = ParserRuleContext{value: "array-length"}
	PARSER_RULE_CONTEXT_ARRAY_LENGTH_START                         = ParserRuleContext{value: "array-length-start"}
	PARSER_RULE_CONTEXT_ANNOT_REFERENCE                            = ParserRuleContext{value: "annot-reference"}
	PARSER_RULE_CONTEXT_ANNOTATIONS                                = ParserRuleContext{value: "annots"}
	PARSER_RULE_CONTEXT_ANNOTATION_END                             = ParserRuleContext{value: "annot-end"}
	PARSER_RULE_CONTEXT_ANNOTATION_REF_RHS                         = ParserRuleContext{value: "annot-ref-rhs"}
	PARSER_RULE_CONTEXT_DOC_STRING                                 = ParserRuleContext{value: "doc-string"}
	PARSER_RULE_CONTEXT_QUALIFIED_IDENTIFIER                       = ParserRuleContext{value: "qualified-identifier"}
	PARSER_RULE_CONTEXT_EQUAL_OR_RIGHT_ARROW                       = ParserRuleContext{value: "equal-or-right-arrow"}
	PARSER_RULE_CONTEXT_ANNOTATION_DECL                            = ParserRuleContext{value: "annotation-decl"}
	PARSER_RULE_CONTEXT_ANNOT_DECL_OPTIONAL_TYPE                   = ParserRuleContext{value: "annot-decl-optional-type"}
	PARSER_RULE_CONTEXT_ANNOT_DECL_RHS                             = ParserRuleContext{value: "annot-decl-rhs"}
	PARSER_RULE_CONTEXT_ANNOT_OPTIONAL_ATTACH_POINTS               = ParserRuleContext{value: "annot-optional-attach-points"}
	PARSER_RULE_CONTEXT_ANNOT_ATTACH_POINTS_LIST                   = ParserRuleContext{value: "annot-attach-points-list"}
	PARSER_RULE_CONTEXT_ATTACH_POINT                               = ParserRuleContext{value: "attach-point"}
	PARSER_RULE_CONTEXT_ATTACH_POINT_IDENT                         = ParserRuleContext{value: "attach-point-ident"}
	PARSER_RULE_CONTEXT_SINGLE_KEYWORD_ATTACH_POINT_IDENT          = ParserRuleContext{value: "single-keyword-attach-point-ident"}
	PARSER_RULE_CONTEXT_IDENT_AFTER_OBJECT_IDENT                   = ParserRuleContext{value: "ident-after-object-ident"}
	PARSER_RULE_CONTEXT_XML_NAMESPACE_DECLARATION                  = ParserRuleContext{value: "xml-namespace-decl"}
	PARSER_RULE_CONTEXT_XML_NAMESPACE_PREFIX_DECL                  = ParserRuleContext{value: "namespace-prefix-decl"}
	PARSER_RULE_CONTEXT_DEFAULT_WORKER_INIT                        = ParserRuleContext{value: "default-worker-init"}
	PARSER_RULE_CONTEXT_NAMED_WORKERS                              = ParserRuleContext{value: "named-workers"}
	PARSER_RULE_CONTEXT_WORKER_NAME_RHS                            = ParserRuleContext{value: "worker-name-rhs"}
	PARSER_RULE_CONTEXT_DEFAULT_WORKER                             = ParserRuleContext{value: "default-worker-init"}
	PARSER_RULE_CONTEXT_KEY_SPECIFIER                              = ParserRuleContext{value: "key-specifier"}
	PARSER_RULE_CONTEXT_KEY_SPECIFIER_RHS                          = ParserRuleContext{value: "key-specifier-rhs"}
	PARSER_RULE_CONTEXT_TABLE_KEY_RHS                              = ParserRuleContext{value: "table-key-rhs"}
	PARSER_RULE_CONTEXT_LET_EXPR_LET_VAR_DECL                      = ParserRuleContext{value: "let-expr-let-var-decl"}
	PARSER_RULE_CONTEXT_LET_CLAUSE_LET_VAR_DECL                    = ParserRuleContext{value: "let-clause-let-var-decl"}
	PARSER_RULE_CONTEXT_LET_VAR_DECL_START                         = ParserRuleContext{value: "let-var-decl-start"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC                             = ParserRuleContext{value: "func-type-desc"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_START                       = ParserRuleContext{value: "func-type-desc-start"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_FIRST_QUALIFIER                  = ParserRuleContext{value: "func-type-first-qualifier"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_SECOND_QUALIFIER                 = ParserRuleContext{value: "func-type-second-qualifier"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL    = ParserRuleContext{value: "func-type-desc-start-without-first-qual"}
	PARSER_RULE_CONTEXT_FUNCTION_KEYWORD_RHS                       = ParserRuleContext{value: "func-keyword-rhs"}
	PARSER_RULE_CONTEXT_END_OF_TYPE_DESC                           = ParserRuleContext{value: "end-of-type-desc"}
	PARSER_RULE_CONTEXT_SELECT_CLAUSE                              = ParserRuleContext{value: "select-clause"}
	PARSER_RULE_CONTEXT_COLLECT_CLAUSE                             = ParserRuleContext{value: "collect-clause"}
	PARSER_RULE_CONTEXT_RESULT_CLAUSE                              = ParserRuleContext{value: "result-clause"}
	PARSER_RULE_CONTEXT_WHERE_CLAUSE                               = ParserRuleContext{value: "where-clause"}
	PARSER_RULE_CONTEXT_FROM_CLAUSE                                = ParserRuleContext{value: "from-clause"}
	PARSER_RULE_CONTEXT_LET_CLAUSE                                 = ParserRuleContext{value: "let-clause"}
	PARSER_RULE_CONTEXT_MODULE_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS  = ParserRuleContext{value: "module-level-func-type-desc-rhs"}
	PARSER_RULE_CONTEXT_EXPLICIT_ANON_FUNC_EXPR_BODY_START         = ParserRuleContext{value: "explicit-anon-func-expr-body-start"}
	PARSER_RULE_CONTEXT_BRACED_EXPR_OR_ANON_FUNC_PARAMS            = ParserRuleContext{value: "braced-expr-or-anon-func-params"}
	PARSER_RULE_CONTEXT_BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS         = ParserRuleContext{value: "braced-expr-or-anon-func-param-rhs"}
	PARSER_RULE_CONTEXT_ANON_FUNC_PARAM_RHS                        = ParserRuleContext{value: "anon-func-param-rhs"}
	PARSER_RULE_CONTEXT_IMPLICIT_ANON_FUNC_PARAM                   = ParserRuleContext{value: "implicit-anon-func-param"}
	PARSER_RULE_CONTEXT_OPTIONAL_PEER_WORKER                       = ParserRuleContext{value: "optional-peer-worker"}
	PARSER_RULE_CONTEXT_METHOD_NAME                                = ParserRuleContext{value: "method-name"}
	PARSER_RULE_CONTEXT_PEER_WORKER_NAME                           = ParserRuleContext{value: "peer-worker-name"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_TUPLE_RHS                     = ParserRuleContext{value: "type-desc-in-tuple-rhs"}
	PARSER_RULE_CONTEXT_TUPLE_TYPE_MEMBER_RHS                      = ParserRuleContext{value: "tuple-type-member-rhs"}
	PARSER_RULE_CONTEXT_NIL_OR_PARENTHESISED_TYPE_DESC_RHS         = ParserRuleContext{value: "nil-or-parenthesised-tpe-desc-rhs"}
	PARSER_RULE_CONTEXT_REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS  = ParserRuleContext{value: "remote-or-resource-call-or-async-send-rhs"}
	PARSER_RULE_CONTEXT_REMOTE_CALL_OR_ASYNC_SEND_END              = ParserRuleContext{value: "remote-call-or-async-send-end"}
	PARSER_RULE_CONTEXT_DEFAULT_WORKER_NAME_IN_ASYNC_SEND          = ParserRuleContext{value: "default-worker-name-in-async-send"}
	PARSER_RULE_CONTEXT_RECEIVE_WORKERS                            = ParserRuleContext{value: "receive-workers"}
	PARSER_RULE_CONTEXT_MULTI_RECEIVE_WORKERS                      = ParserRuleContext{value: "multi-receive-workers"}
	PARSER_RULE_CONTEXT_RECEIVE_FIELD_END                          = ParserRuleContext{value: "receive-field-end"}
	PARSER_RULE_CONTEXT_RECEIVE_FIELD                              = ParserRuleContext{value: "receive-field"}
	PARSER_RULE_CONTEXT_RECEIVE_FIELD_NAME                         = ParserRuleContext{value: "receive-field-name"}
	PARSER_RULE_CONTEXT_INFER_PARAM_END_OR_PARENTHESIS_END         = ParserRuleContext{value: "infer-param-end-or-parenthesis-end"}
	PARSER_RULE_CONTEXT_LIST_CONSTRUCTOR_MEMBER_END                = ParserRuleContext{value: "list-constructor-member-end"}
	PARSER_RULE_CONTEXT_TYPED_BINDING_PATTERN                      = ParserRuleContext{value: "typed-binding-pattern"}
	PARSER_RULE_CONTEXT_BINDING_PATTERN                            = ParserRuleContext{value: "binding-pattern"}
	PARSER_RULE_CONTEXT_CAPTURE_BINDING_PATTERN                    = ParserRuleContext{value: "capture-binding-pattern"}
	PARSER_RULE_CONTEXT_REST_BINDING_PATTERN                       = ParserRuleContext{value: "rest-binding-pattern"}
	PARSER_RULE_CONTEXT_LIST_BINDING_PATTERN                       = ParserRuleContext{value: "list-binding-pattern"}
	PARSER_RULE_CONTEXT_LIST_BINDING_PATTERNS_START                = ParserRuleContext{value: "list-binding-patterns-start"}
	PARSER_RULE_CONTEXT_LIST_BINDING_PATTERN_MEMBER                = ParserRuleContext{value: "list-binding-pattern-member"}
	PARSER_RULE_CONTEXT_LIST_BINDING_PATTERN_MEMBER_END            = ParserRuleContext{value: "list-binding-pattern-member-end"}
	PARSER_RULE_CONTEXT_FIELD_BINDING_PATTERN                      = ParserRuleContext{value: "field-binding-pattern"}
	PARSER_RULE_CONTEXT_FIELD_BINDING_PATTERN_NAME                 = ParserRuleContext{value: "field-binding-pattern-name"}
	PARSER_RULE_CONTEXT_MAPPING_BINDING_PATTERN                    = ParserRuleContext{value: "mapping-binding-pattern"}
	PARSER_RULE_CONTEXT_MAPPING_BINDING_PATTERN_MEMBER             = ParserRuleContext{value: "mapping-binding-pattern-member"}
	PARSER_RULE_CONTEXT_MAPPING_BINDING_PATTERN_END                = ParserRuleContext{value: "mapping-binding-pattern-end"}
	PARSER_RULE_CONTEXT_FIELD_BINDING_PATTERN_END                  = ParserRuleContext{value: "field-binding-pattern-end-or-continue"}
	PARSER_RULE_CONTEXT_ERROR_BINDING_PATTERN                      = ParserRuleContext{value: "error-binding-pattern"}
	PARSER_RULE_CONTEXT_ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS    = ParserRuleContext{value: "error-binding-pattern-error-keyword-rhs"}
	PARSER_RULE_CONTEXT_ERROR_ARG_LIST_BINDING_PATTERN_START       = ParserRuleContext{value: "error-arg-list-binding-pattern-start"}
	PARSER_RULE_CONTEXT_SIMPLE_BINDING_PATTERN                     = ParserRuleContext{value: "simple-binding-pattern"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_BINDING_PATTERN_END          = ParserRuleContext{value: "error-message-binding-pattern-end"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_BINDING_PATTERN_END_COMMA    = ParserRuleContext{value: "error-message-binding-pattern-end-comma"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_BINDING_PATTERN_RHS          = ParserRuleContext{value: "error-message-binding-pattern-rhs"}
	PARSER_RULE_CONTEXT_ERROR_CAUSE_SIMPLE_BINDING_PATTERN         = ParserRuleContext{value: "error-cause-simple-binding-pattern"}
	PARSER_RULE_CONTEXT_ERROR_FIELD_BINDING_PATTERN                = ParserRuleContext{value: "error-field-binding-pattern"}
	PARSER_RULE_CONTEXT_ERROR_FIELD_BINDING_PATTERN_END            = ParserRuleContext{value: "error-field-binding-pattern-end"}
	PARSER_RULE_CONTEXT_NAMED_ARG_BINDING_PATTERN                  = ParserRuleContext{value: "named-arg-binding-pattern"}
	PARSER_RULE_CONTEXT_BINDING_PATTERN_STARTING_IDENTIFIER        = ParserRuleContext{value: "binding-pattern-starting-indentifier"}
	PARSER_RULE_CONTEXT_WAIT_KEYWORD_RHS                           = ParserRuleContext{value: "wait-keyword-rhs"}
	PARSER_RULE_CONTEXT_MULTI_WAIT_FIELDS                          = ParserRuleContext{value: "multi-wait-fields"}
	PARSER_RULE_CONTEXT_WAIT_FIELD_NAME                            = ParserRuleContext{value: "wait-field-name"}
	PARSER_RULE_CONTEXT_WAIT_FIELD_NAME_RHS                        = ParserRuleContext{value: "wait-field-name-rhs"}
	PARSER_RULE_CONTEXT_WAIT_FIELD_END                             = ParserRuleContext{value: "wait-field-end"}
	PARSER_RULE_CONTEXT_WAIT_FUTURE_EXPR_END                       = ParserRuleContext{value: "wait-future-expr-end"}
	PARSER_RULE_CONTEXT_ALTERNATE_WAIT_EXPRS                       = ParserRuleContext{value: "alternate-wait-exprs"}
	PARSER_RULE_CONTEXT_ALTERNATE_WAIT_EXPR_LIST_END               = ParserRuleContext{value: "alternate-wait-expr-lit-end"}
	PARSER_RULE_CONTEXT_DO_CLAUSE                                  = ParserRuleContext{value: "do-clause"}
	PARSER_RULE_CONTEXT_MODULE_ENUM_DECLARATION                    = ParserRuleContext{value: "module-enum-declaration"}
	PARSER_RULE_CONTEXT_MODULE_ENUM_NAME                           = ParserRuleContext{value: "module-enum-name"}
	PARSER_RULE_CONTEXT_ENUM_MEMBER_NAME                           = ParserRuleContext{value: "enum-member-name"}
	PARSER_RULE_CONTEXT_MEMBER_ACCESS_KEY_EXPR_END                 = ParserRuleContext{value: "member-access-key-expr-end"}
	PARSER_RULE_CONTEXT_MEMBER_ACCESS_KEY_EXPR                     = ParserRuleContext{value: "member-access-key-expr"}
	PARSER_RULE_CONTEXT_RETRY_KEYWORD_RHS                          = ParserRuleContext{value: "retry-keyword-rhs"}
	PARSER_RULE_CONTEXT_RETRY_TYPE_PARAM_RHS                       = ParserRuleContext{value: "retry-type-param-rhs"}
	PARSER_RULE_CONTEXT_RETRY_BODY                                 = ParserRuleContext{value: "retry-body"}
	PARSER_RULE_CONTEXT_ROLLBACK_RHS                               = ParserRuleContext{value: "rollback-rhs"}
	PARSER_RULE_CONTEXT_STMT_START_BRACKETED_LIST                  = ParserRuleContext{value: "stmt-start-bracketed-list"}
	PARSER_RULE_CONTEXT_STMT_START_BRACKETED_LIST_MEMBER           = ParserRuleContext{value: "stmt-start-bracketed-list-member"}
	PARSER_RULE_CONTEXT_STMT_START_BRACKETED_LIST_RHS              = ParserRuleContext{value: "stmt-start-bracketed-list-rhs"}
	PARSER_RULE_CONTEXT_BRACKETED_LIST                             = ParserRuleContext{value: "bracketed-list"}
	PARSER_RULE_CONTEXT_BRACKETED_LIST_RHS                         = ParserRuleContext{value: "bracketed-list-rhs"}
	PARSER_RULE_CONTEXT_BRACED_LIST_RHS                            = ParserRuleContext{value: "braced-list-rhs"}
	PARSER_RULE_CONTEXT_BRACKETED_LIST_MEMBER                      = ParserRuleContext{value: "bracketed-list-member"}
	PARSER_RULE_CONTEXT_BRACKETED_LIST_MEMBER_END                  = ParserRuleContext{value: "bracketed-list-member-end"}
	PARSER_RULE_CONTEXT_LIST_BINDING_MEMBER_OR_ARRAY_LENGTH        = ParserRuleContext{value: "list-binding-member-or-array-length"}
	PARSER_RULE_CONTEXT_TYPED_BINDING_PATTERN_TYPE_RHS             = ParserRuleContext{value: "type-binding-pattern-type-rhs"}
	PARSER_RULE_CONTEXT_UNION_OR_INTERSECTION_TOKEN                = ParserRuleContext{value: "union-or-intersection"}
	PARSER_RULE_CONTEXT_MAPPING_BP_OR_MAPPING_CONSTRUCTOR          = ParserRuleContext{value: "mapping-bp-or-mapping-cons"}
	PARSER_RULE_CONTEXT_MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER   = ParserRuleContext{value: "mapping-bp-or-mapping-cons-member"}
	PARSER_RULE_CONTEXT_LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER         = ParserRuleContext{value: "list-bp-or-list-cons-member"}
	PARSER_RULE_CONTEXT_VAR_REF_OR_TYPE_REF                        = ParserRuleContext{value: "var-ref"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_OR_ANON_FUNC                = ParserRuleContext{value: "func-desc-type-or-anon-func"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_OR_ANON_FUNC_START          = ParserRuleContext{value: "func-desc-type-or-anon-func-start"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY       = ParserRuleContext{value: "func-type-desc-rhs-or-anon-func-body"}
	PARSER_RULE_CONTEXT_STMT_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS    = ParserRuleContext{value: "stmt-level-func-type-desc-rhs"}
	PARSER_RULE_CONTEXT_RECORD_FIELD_NAME_OR_TYPE_NAME             = ParserRuleContext{value: "record-field-name-or-type-name"}
	PARSER_RULE_CONTEXT_MATCH_BODY                                 = ParserRuleContext{value: "match-body"}
	PARSER_RULE_CONTEXT_MATCH_PATTERN                              = ParserRuleContext{value: "match-pattern"}
	PARSER_RULE_CONTEXT_MATCH_PATTERN_START                        = ParserRuleContext{value: "match-pattern-start"}
	PARSER_RULE_CONTEXT_MATCH_PATTERN_END                          = ParserRuleContext{value: "match-pattern-end"}
	PARSER_RULE_CONTEXT_MATCH_PATTERN_RHS                          = ParserRuleContext{value: "match-pattern-rhs"}
	PARSER_RULE_CONTEXT_MATCH_PATTERN_LIST_MEMBER_RHS              = ParserRuleContext{value: "match-pattern-list-memebr-rhs"}
	PARSER_RULE_CONTEXT_OPTIONAL_MATCH_GUARD                       = ParserRuleContext{value: "optional-match-guard"}
	PARSER_RULE_CONTEXT_LIST_MATCH_PATTERN                         = ParserRuleContext{value: "list-match-pattern"}
	PARSER_RULE_CONTEXT_LIST_MATCH_PATTERNS_START                  = ParserRuleContext{value: "list-match-patterns-start"}
	PARSER_RULE_CONTEXT_LIST_MATCH_PATTERN_MEMBER                  = ParserRuleContext{value: "list-match-pattern-member"}
	PARSER_RULE_CONTEXT_LIST_MATCH_PATTERN_MEMBER_RHS              = ParserRuleContext{value: "list-match-pattern-member-rhs"}
	PARSER_RULE_CONTEXT_REST_MATCH_PATTERN                         = ParserRuleContext{value: "rest-match-pattern"}
	PARSER_RULE_CONTEXT_MAPPING_MATCH_PATTERN                      = ParserRuleContext{value: "mapping-match-pattern"}
	PARSER_RULE_CONTEXT_FIELD_MATCH_PATTERNS_START                 = ParserRuleContext{value: "field-match-patterns-start"}
	PARSER_RULE_CONTEXT_FIELD_MATCH_PATTERN_MEMBER_RHS             = ParserRuleContext{value: "field-match-pattern-member-rhs"}
	PARSER_RULE_CONTEXT_FIELD_MATCH_PATTERN_MEMBER                 = ParserRuleContext{value: "field-match-pattern-member"}
	PARSER_RULE_CONTEXT_ERROR_MATCH_PATTERN                        = ParserRuleContext{value: "error-match-pattern"}
	PARSER_RULE_CONTEXT_ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS      = ParserRuleContext{value: "error-match-pattern-error-keyword-rhs"}
	PARSER_RULE_CONTEXT_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG     = ParserRuleContext{value: "error-arg-list-match-pattern-first-arg"}
	PARSER_RULE_CONTEXT_ERROR_ARG_LIST_MATCH_PATTERN_START         = ParserRuleContext{value: "error-arg-list-match-pattern-start"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_MATCH_PATTERN_END            = ParserRuleContext{value: "error-message-match-pattern-end"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_MATCH_PATTERN_END_COMMA      = ParserRuleContext{value: "error-message-match-pattern-end-comma"}
	PARSER_RULE_CONTEXT_ERROR_MESSAGE_MATCH_PATTERN_RHS            = ParserRuleContext{value: "error-message-match-pattern-rhs"}
	PARSER_RULE_CONTEXT_ERROR_CAUSE_MATCH_PATTERN                  = ParserRuleContext{value: "error-cause-match-pattern"}
	PARSER_RULE_CONTEXT_ERROR_FIELD_MATCH_PATTERN                  = ParserRuleContext{value: "error-field-match-pattern"}
	PARSER_RULE_CONTEXT_ERROR_FIELD_MATCH_PATTERN_RHS              = ParserRuleContext{value: "error-field-match-pattern-rhs"}
	PARSER_RULE_CONTEXT_ERROR_MATCH_PATTERN_OR_CONST_PATTERN       = ParserRuleContext{value: "error-match-pattern-or-const-pattern"}
	PARSER_RULE_CONTEXT_NAMED_ARG_MATCH_PATTERN                    = ParserRuleContext{value: "named-arg-match-pattern"}
	PARSER_RULE_CONTEXT_NAMED_ARG_MATCH_PATTERN_RHS                = ParserRuleContext{value: "named-arg-match-pattern-rhs"}
	PARSER_RULE_CONTEXT_ORDER_BY_CLAUSE                            = ParserRuleContext{value: "order-by-clause"}
	PARSER_RULE_CONTEXT_ORDER_KEY_LIST                             = ParserRuleContext{value: "order-key-list"}
	PARSER_RULE_CONTEXT_ORDER_KEY_LIST_END                         = ParserRuleContext{value: "order-key-list-end"}
	PARSER_RULE_CONTEXT_GROUP_BY_CLAUSE                            = ParserRuleContext{value: "group-by-clause"}
	PARSER_RULE_CONTEXT_GROUPING_KEY_LIST_ELEMENT                  = ParserRuleContext{value: "grouping-key-list-element"}
	PARSER_RULE_CONTEXT_GROUPING_KEY_LIST_ELEMENT_END              = ParserRuleContext{value: "grouping-key-list-element-end"}
	PARSER_RULE_CONTEXT_GROUP_BY_CLAUSE_END                        = ParserRuleContext{value: "group-by-clause-end"}
	PARSER_RULE_CONTEXT_ON_CONFLICT_CLAUSE                         = ParserRuleContext{value: "on-conflict-clause"}
	PARSER_RULE_CONTEXT_LIMIT_CLAUSE                               = ParserRuleContext{value: "limit-clause"}
	PARSER_RULE_CONTEXT_JOIN_CLAUSE                                = ParserRuleContext{value: "join-clause"}
	PARSER_RULE_CONTEXT_JOIN_CLAUSE_START                          = ParserRuleContext{value: "join-clause-start"}
	PARSER_RULE_CONTEXT_JOIN_CLAUSE_END                            = ParserRuleContext{value: "join-clause-end"}
	PARSER_RULE_CONTEXT_ON_CLAUSE                                  = ParserRuleContext{value: "on-clause"}
	PARSER_RULE_CONTEXT_INTERMEDIATE_CLAUSE                        = ParserRuleContext{value: "intermediate-clause"}
	PARSER_RULE_CONTEXT_INTERMEDIATE_CLAUSE_START                  = ParserRuleContext{value: "intermediate-clause-start"}
	PARSER_RULE_CONTEXT_ON_FAIL_CLAUSE                             = ParserRuleContext{value: "on_fail_clause"}
	PARSER_RULE_CONTEXT_ON_FA                                      = ParserRuleContext{value: "on_fail_clause"}
	PARSER_RULE_CONTEXT_OPTIONAL_TYPE_PARAMETER                    = ParserRuleContext{value: "optional-type-parameter"}
	PARSER_RULE_CONTEXT_PARAMETERIZED_TYPE                         = ParserRuleContext{value: "parameterized-type"}
	PARSER_RULE_CONTEXT_MAP_TYPE_DESCRIPTOR                        = ParserRuleContext{value: "map-type-descriptor"}
	PARSER_RULE_CONTEXT_MODULE_VAR_DECL                            = ParserRuleContext{value: "module-var-decl"}
	PARSER_RULE_CONTEXT_MODULE_VAR_FIRST_QUAL                      = ParserRuleContext{value: "module-var-first-qual"}
	PARSER_RULE_CONTEXT_MODULE_VAR_SECOND_QUAL                     = ParserRuleContext{value: "module-var-second-qual"}
	PARSER_RULE_CONTEXT_MODULE_VAR_THIRD_QUAL                      = ParserRuleContext{value: "module-var-third-qual"}
	PARSER_RULE_CONTEXT_MODULE_VAR_DECL_START                      = ParserRuleContext{value: "module-var-decl-start"}
	PARSER_RULE_CONTEXT_MODULE_VAR_WITHOUT_FIRST_QUAL              = ParserRuleContext{value: "module-var-without-first-qual"}
	PARSER_RULE_CONTEXT_MODULE_VAR_WITHOUT_SECOND_QUAL             = ParserRuleContext{value: "module-var-without-second-qual"}
	PARSER_RULE_CONTEXT_FUNC_DEF_OR_TYPE_DESC_RHS                  = ParserRuleContext{value: "func-def-or-type-desc-rhs"}
	PARSER_RULE_CONTEXT_CLIENT_RESOURCE_ACCESS_ACTION              = ParserRuleContext{value: "client-resource-access-action"}
	PARSER_RULE_CONTEXT_OPTIONAL_RESOURCE_ACCESS_PATH              = ParserRuleContext{value: "optional-resource-access-path"}
	PARSER_RULE_CONTEXT_RESOURCE_ACCESS_PATH_SEGMENT               = ParserRuleContext{value: "resource-access-path-segment"}
	PARSER_RULE_CONTEXT_COMPUTED_SEGMENT_OR_REST_SEGMENT           = ParserRuleContext{value: "computed-segment-or-rest-segment"}
	PARSER_RULE_CONTEXT_RESOURCE_ACCESS_SEGMENT_RHS                = ParserRuleContext{value: "resource-access-segment-rhs"}
	PARSER_RULE_CONTEXT_OPTIONAL_RESOURCE_ACCESS_METHOD            = ParserRuleContext{value: "optional-resource-access-method"}
	PARSER_RULE_CONTEXT_OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST   = ParserRuleContext{value: "optional-resource-method-call-arg-list"}
	PARSER_RULE_CONTEXT_ACTION_END                                 = ParserRuleContext{value: "action-end"}
	PARSER_RULE_CONTEXT_OPTIONAL_PARENTHESIZED_ARG_LIST            = ParserRuleContext{value: "optional-parenthesized-arg-list"}
	PARSER_RULE_CONTEXT_NATURAL_EXPRESSION                         = ParserRuleContext{value: "natural-expression"}
	PARSER_RULE_CONTEXT_NATURAL_EXPRESSION_START                   = ParserRuleContext{value: "natural-expression-start"}

	// Statements
	PARSER_RULE_CONTEXT_STATEMENT                      = ParserRuleContext{value: "statement"}
	PARSER_RULE_CONTEXT_STATEMENTS                     = ParserRuleContext{value: "statements"}
	PARSER_RULE_CONTEXT_STATEMENT_WITHOUT_ANNOTS       = ParserRuleContext{value: "statement-without-annots"}
	PARSER_RULE_CONTEXT_ASSIGNMENT_STMT                = ParserRuleContext{value: "assignment-stmt"}
	PARSER_RULE_CONTEXT_VAR_DECL_STMT                  = ParserRuleContext{value: "var-decl-stmt"}
	PARSER_RULE_CONTEXT_VAR_DECL_STMT_RHS              = ParserRuleContext{value: "var-decl-rhs"}
	PARSER_RULE_CONTEXT_CONFIG_VAR_DECL_RHS            = ParserRuleContext{value: "config-var-decl-rhs"}
	PARSER_RULE_CONTEXT_TYPE_NAME_OR_VAR_NAME          = ParserRuleContext{value: "type-or-var-name"}
	PARSER_RULE_CONTEXT_ASSIGNMENT_OR_VAR_DECL_STMT    = ParserRuleContext{value: "assign-or-var-decl"}
	PARSER_RULE_CONTEXT_IF_BLOCK                       = ParserRuleContext{value: "if-block"}
	PARSER_RULE_CONTEXT_BLOCK_STMT                     = ParserRuleContext{value: "block-stmt"}
	PARSER_RULE_CONTEXT_ELSE_BLOCK                     = ParserRuleContext{value: "else-block"}
	PARSER_RULE_CONTEXT_ELSE_BODY                      = ParserRuleContext{value: "else-body"}
	PARSER_RULE_CONTEXT_WHILE_BLOCK                    = ParserRuleContext{value: "while-block"}
	PARSER_RULE_CONTEXT_DO_BLOCK                       = ParserRuleContext{value: "do-block"}
	PARSER_RULE_CONTEXT_CALL_STMT                      = ParserRuleContext{value: "call-statement"}
	PARSER_RULE_CONTEXT_CALL_STMT_START                = ParserRuleContext{value: "call-statement-start"}
	PARSER_RULE_CONTEXT_CONTINUE_STATEMENT             = ParserRuleContext{value: "continue-statement"}
	PARSER_RULE_CONTEXT_BREAK_STATEMENT                = ParserRuleContext{value: "break-statement"}
	PARSER_RULE_CONTEXT_PANIC_STMT                     = ParserRuleContext{value: "panic-statement"}
	PARSER_RULE_CONTEXT_RETURN_STMT                    = ParserRuleContext{value: "return-stmt"}
	PARSER_RULE_CONTEXT_RETURN_STMT_RHS                = ParserRuleContext{value: "return-stmt-rhs"}
	PARSER_RULE_CONTEXT_REGULAR_COMPOUND_STMT_RHS      = ParserRuleContext{value: "regular-compound-statement-rhs"}
	PARSER_RULE_CONTEXT_LOCAL_TYPE_DEFINITION_STMT     = ParserRuleContext{value: "local-type-definition-statement"}
	PARSER_RULE_CONTEXT_BINDING_PATTERN_OR_EXPR_RHS    = ParserRuleContext{value: "binding-pattern-or-expr-rhs"}
	PARSER_RULE_CONTEXT_BINDING_PATTERN_OR_VAR_REF_RHS = ParserRuleContext{value: "binding.pattern.or.var.ref.rhs"}
	PARSER_RULE_CONTEXT_TYPE_DESC_OR_EXPR_RHS          = ParserRuleContext{value: "type-desc-or-expr-rhs"}
	PARSER_RULE_CONTEXT_STMT_START_WITH_EXPR_RHS       = ParserRuleContext{value: "stmt-start-with-expr-rhs"}
	PARSER_RULE_CONTEXT_EXPR_STMT_RHS                  = ParserRuleContext{value: "expr-stmt-rhs"}
	PARSER_RULE_CONTEXT_EXPRESSION_STATEMENT           = ParserRuleContext{value: "expression-statement"}
	PARSER_RULE_CONTEXT_EXPRESSION_STATEMENT_START     = ParserRuleContext{value: "expression-statement-start"}
	PARSER_RULE_CONTEXT_LOCK_STMT                      = ParserRuleContext{value: "lock-stmt"}
	PARSER_RULE_CONTEXT_NAMED_WORKER_DECL              = ParserRuleContext{value: "named-worker-decl"}
	PARSER_RULE_CONTEXT_NAMED_WORKER_DECL_START        = ParserRuleContext{value: "named-worker-decl-start"}
	PARSER_RULE_CONTEXT_FORK_STMT                      = ParserRuleContext{value: "fork-stmt"}
	PARSER_RULE_CONTEXT_FOREACH_STMT                   = ParserRuleContext{value: "foreach-stmt"}
	PARSER_RULE_CONTEXT_TRANSACTION_STMT               = ParserRuleContext{value: "transaction-stmt"}
	PARSER_RULE_CONTEXT_RETRY_STMT                     = ParserRuleContext{value: "retry-stmt"}
	PARSER_RULE_CONTEXT_ROLLBACK_STMT                  = ParserRuleContext{value: "rollback-stmt"}
	PARSER_RULE_CONTEXT_AMBIGUOUS_STMT                 = ParserRuleContext{value: "ambiguous-stmt"}
	PARSER_RULE_CONTEXT_MATCH_STMT                     = ParserRuleContext{value: "match-stmt"}
	PARSER_RULE_CONTEXT_FAIL_STATEMENT                 = ParserRuleContext{value: "fail-stmt"}

	// Keywords
	PARSER_RULE_CONTEXT_RETURNS_KEYWORD       = ParserRuleContext{value: "returns"}
	PARSER_RULE_CONTEXT_TYPE_KEYWORD          = ParserRuleContext{value: "type"}
	PARSER_RULE_CONTEXT_CLASS_KEYWORD         = ParserRuleContext{value: "class"}
	PARSER_RULE_CONTEXT_PUBLIC_KEYWORD        = ParserRuleContext{value: "public"}
	PARSER_RULE_CONTEXT_PRIVATE_KEYWORD       = ParserRuleContext{value: "private"}
	PARSER_RULE_CONTEXT_FUNCTION_KEYWORD      = ParserRuleContext{value: "function"}
	PARSER_RULE_CONTEXT_EXTERNAL_KEYWORD      = ParserRuleContext{value: "external"}
	PARSER_RULE_CONTEXT_RECORD_KEYWORD        = ParserRuleContext{value: "record"}
	PARSER_RULE_CONTEXT_OBJECT_KEYWORD        = ParserRuleContext{value: "object"}
	PARSER_RULE_CONTEXT_ABSTRACT_KEYWORD      = ParserRuleContext{value: "abstract"}
	PARSER_RULE_CONTEXT_CLIENT_KEYWORD        = ParserRuleContext{value: "client"}
	PARSER_RULE_CONTEXT_IF_KEYWORD            = ParserRuleContext{value: "if"}
	PARSER_RULE_CONTEXT_ELSE_KEYWORD          = ParserRuleContext{value: "else"}
	PARSER_RULE_CONTEXT_WHILE_KEYWORD         = ParserRuleContext{value: "while"}
	PARSER_RULE_CONTEXT_CONTINUE_KEYWORD      = ParserRuleContext{value: "continue"}
	PARSER_RULE_CONTEXT_BREAK_KEYWORD         = ParserRuleContext{value: "break"}
	PARSER_RULE_CONTEXT_PANIC_KEYWORD         = ParserRuleContext{value: "panic"}
	PARSER_RULE_CONTEXT_IMPORT_KEYWORD        = ParserRuleContext{value: "import"}
	PARSER_RULE_CONTEXT_AS_KEYWORD            = ParserRuleContext{value: "as"}
	PARSER_RULE_CONTEXT_RETURN_KEYWORD        = ParserRuleContext{value: "return"}
	PARSER_RULE_CONTEXT_SERVICE_KEYWORD       = ParserRuleContext{value: "service"}
	PARSER_RULE_CONTEXT_ON_KEYWORD            = ParserRuleContext{value: "on"}
	PARSER_RULE_CONTEXT_FINAL_KEYWORD         = ParserRuleContext{value: "final"}
	PARSER_RULE_CONTEXT_LISTENER_KEYWORD      = ParserRuleContext{value: "listener"}
	PARSER_RULE_CONTEXT_CONST_KEYWORD         = ParserRuleContext{value: "const"}
	PARSER_RULE_CONTEXT_TYPEOF_KEYWORD        = ParserRuleContext{value: "typeof"}
	PARSER_RULE_CONTEXT_IS_KEYWORD            = ParserRuleContext{value: "is"}
	PARSER_RULE_CONTEXT_MAP_KEYWORD           = ParserRuleContext{value: "map"}
	PARSER_RULE_CONTEXT_NULL_KEYWORD          = ParserRuleContext{value: "null"}
	PARSER_RULE_CONTEXT_LOCK_KEYWORD          = ParserRuleContext{value: "lock"}
	PARSER_RULE_CONTEXT_ANNOTATION_KEYWORD    = ParserRuleContext{value: "annotation"}
	PARSER_RULE_CONTEXT_SOURCE_KEYWORD        = ParserRuleContext{value: "source"}
	PARSER_RULE_CONTEXT_XMLNS_KEYWORD         = ParserRuleContext{value: "xmlns"}
	PARSER_RULE_CONTEXT_WORKER_KEYWORD        = ParserRuleContext{value: "worker"}
	PARSER_RULE_CONTEXT_FORK_KEYWORD          = ParserRuleContext{value: "fork"}
	PARSER_RULE_CONTEXT_TRAP_KEYWORD          = ParserRuleContext{value: "trap"}
	PARSER_RULE_CONTEXT_IN_KEYWORD            = ParserRuleContext{value: "in"}
	PARSER_RULE_CONTEXT_FOREACH_KEYWORD       = ParserRuleContext{value: "foreach"}
	PARSER_RULE_CONTEXT_TABLE_KEYWORD         = ParserRuleContext{value: "table"}
	PARSER_RULE_CONTEXT_KEY_KEYWORD           = ParserRuleContext{value: "key"}
	PARSER_RULE_CONTEXT_ERROR_KEYWORD         = ParserRuleContext{value: "error"}
	PARSER_RULE_CONTEXT_LET_KEYWORD           = ParserRuleContext{value: "let"}
	PARSER_RULE_CONTEXT_STREAM_KEYWORD        = ParserRuleContext{value: "stream"}
	PARSER_RULE_CONTEXT_XML_KEYWORD           = ParserRuleContext{value: "xml"}
	PARSER_RULE_CONTEXT_STRING_KEYWORD        = ParserRuleContext{value: "string"}
	PARSER_RULE_CONTEXT_NEW_KEYWORD           = ParserRuleContext{value: "new"}
	PARSER_RULE_CONTEXT_FROM_KEYWORD          = ParserRuleContext{value: "from"}
	PARSER_RULE_CONTEXT_WHERE_KEYWORD         = ParserRuleContext{value: "where"}
	PARSER_RULE_CONTEXT_SELECT_KEYWORD        = ParserRuleContext{value: "select"}
	PARSER_RULE_CONTEXT_COLLECT_KEYWORD       = ParserRuleContext{value: "collect"}
	PARSER_RULE_CONTEXT_START_KEYWORD         = ParserRuleContext{value: "start"}
	PARSER_RULE_CONTEXT_FLUSH_KEYWORD         = ParserRuleContext{value: "flush"}
	PARSER_RULE_CONTEXT_WAIT_KEYWORD          = ParserRuleContext{value: "wait"}
	PARSER_RULE_CONTEXT_DO_KEYWORD            = ParserRuleContext{value: "do"}
	PARSER_RULE_CONTEXT_TRANSACTION_KEYWORD   = ParserRuleContext{value: "transaction"}
	PARSER_RULE_CONTEXT_COMMIT_KEYWORD        = ParserRuleContext{value: "commit"}
	PARSER_RULE_CONTEXT_RETRY_KEYWORD         = ParserRuleContext{value: "retry"}
	PARSER_RULE_CONTEXT_ROLLBACK_KEYWORD      = ParserRuleContext{value: "rollback"}
	PARSER_RULE_CONTEXT_TRANSACTIONAL_KEYWORD = ParserRuleContext{value: "transactional"}
	PARSER_RULE_CONTEXT_ENUM_KEYWORD          = ParserRuleContext{value: "enum"}
	PARSER_RULE_CONTEXT_BASE16_KEYWORD        = ParserRuleContext{value: "base16"}
	PARSER_RULE_CONTEXT_BASE64_KEYWORD        = ParserRuleContext{value: "base64"}
	PARSER_RULE_CONTEXT_READONLY_KEYWORD      = ParserRuleContext{value: "readonly"}
	PARSER_RULE_CONTEXT_MATCH_KEYWORD         = ParserRuleContext{value: "match"}
	PARSER_RULE_CONTEXT_DISTINCT_KEYWORD      = ParserRuleContext{value: "distinct"}
	PARSER_RULE_CONTEXT_CONFLICT_KEYWORD      = ParserRuleContext{value: "conflict"}
	PARSER_RULE_CONTEXT_LIMIT_KEYWORD         = ParserRuleContext{value: "limit"}
	PARSER_RULE_CONTEXT_JOIN_KEYWORD          = ParserRuleContext{value: "join"}
	PARSER_RULE_CONTEXT_OUTER_KEYWORD         = ParserRuleContext{value: "outer"}
	PARSER_RULE_CONTEXT_VAR_KEYWORD           = ParserRuleContext{value: "var"}
	PARSER_RULE_CONTEXT_FAIL_KEYWORD          = ParserRuleContext{value: "fail"}
	PARSER_RULE_CONTEXT_ORDER_KEYWORD         = ParserRuleContext{value: "order"}
	PARSER_RULE_CONTEXT_BY_KEYWORD            = ParserRuleContext{value: "by"}
	PARSER_RULE_CONTEXT_EQUALS_KEYWORD        = ParserRuleContext{value: "equals"}
	PARSER_RULE_CONTEXT_NOT_IS_KEYWORD        = ParserRuleContext{value: "!is"}
	PARSER_RULE_CONTEXT_RE_KEYWORD            = ParserRuleContext{value: "re"}
	PARSER_RULE_CONTEXT_GROUP_KEYWORD         = ParserRuleContext{value: "group"}
	PARSER_RULE_CONTEXT_NATURAL_KEYWORD       = ParserRuleContext{value: "natural"}

	// Syntax tokens
	PARSER_RULE_CONTEXT_OPEN_PARENTHESIS                      = ParserRuleContext{value: "("}
	PARSER_RULE_CONTEXT_CLOSE_PARENTHESIS                     = ParserRuleContext{value: ")"}
	PARSER_RULE_CONTEXT_OPEN_BRACE                            = ParserRuleContext{value: "{"}
	PARSER_RULE_CONTEXT_CLOSE_BRACE                           = ParserRuleContext{value: "}"}
	PARSER_RULE_CONTEXT_ASSIGN_OP                             = ParserRuleContext{value: "="}
	PARSER_RULE_CONTEXT_SEMICOLON                             = ParserRuleContext{value: ";"}
	PARSER_RULE_CONTEXT_COLON                                 = ParserRuleContext{value: ":"}
	PARSER_RULE_CONTEXT_COMMA                                 = ParserRuleContext{value: ""}
	PARSER_RULE_CONTEXT_ELLIPSIS                              = ParserRuleContext{value: "..."}
	PARSER_RULE_CONTEXT_QUESTION_MARK                         = ParserRuleContext{value: "?"}
	PARSER_RULE_CONTEXT_ASTERISK                              = ParserRuleContext{value: "*"}
	PARSER_RULE_CONTEXT_CLOSED_RECORD_BODY_START              = ParserRuleContext{value: "{|"}
	PARSER_RULE_CONTEXT_CLOSED_RECORD_BODY_END                = ParserRuleContext{value: "|}"}
	PARSER_RULE_CONTEXT_DOT                                   = ParserRuleContext{value: "."}
	PARSER_RULE_CONTEXT_OPEN_BRACKET                          = ParserRuleContext{value: "["}
	PARSER_RULE_CONTEXT_CLOSE_BRACKET                         = ParserRuleContext{value: "]"}
	PARSER_RULE_CONTEXT_SLASH                                 = ParserRuleContext{value: "/"}
	PARSER_RULE_CONTEXT_AT                                    = ParserRuleContext{value: "@"}
	PARSER_RULE_CONTEXT_RIGHT_ARROW                           = ParserRuleContext{value: "->"}
	PARSER_RULE_CONTEXT_GT                                    = ParserRuleContext{value: ">"}
	PARSER_RULE_CONTEXT_LT                                    = ParserRuleContext{value: "<"}
	PARSER_RULE_CONTEXT_PIPE                                  = ParserRuleContext{value: "|"}
	PARSER_RULE_CONTEXT_TEMPLATE_START                        = ParserRuleContext{value: "`"}
	PARSER_RULE_CONTEXT_TEMPLATE_END                          = ParserRuleContext{value: "`"}
	PARSER_RULE_CONTEXT_LT_TOKEN                              = ParserRuleContext{value: "<"}
	PARSER_RULE_CONTEXT_GT_TOKEN                              = ParserRuleContext{value: ">"}
	PARSER_RULE_CONTEXT_ERROR_TYPE_PARAM_START                = ParserRuleContext{value: "<"}
	PARSER_RULE_CONTEXT_PARENTHESISED_TYPE_DESC_START         = ParserRuleContext{value: "("}
	PARSER_RULE_CONTEXT_BITWISE_AND_OPERATOR                  = ParserRuleContext{value: "&"}
	PARSER_RULE_CONTEXT_EXPR_FUNC_BODY_START                  = ParserRuleContext{value: "=>"}
	PARSER_RULE_CONTEXT_PLUS_TOKEN                            = ParserRuleContext{value: "+"}
	PARSER_RULE_CONTEXT_MINUS_TOKEN                           = ParserRuleContext{value: "-"}
	PARSER_RULE_CONTEXT_TUPLE_TYPE_DESC_START                 = ParserRuleContext{value: "["}
	PARSER_RULE_CONTEXT_SYNC_SEND_TOKEN                       = ParserRuleContext{value: "->>"}
	PARSER_RULE_CONTEXT_LEFT_ARROW_TOKEN                      = ParserRuleContext{value: "<-"}
	PARSER_RULE_CONTEXT_ANNOT_CHAINING_TOKEN                  = ParserRuleContext{value: ".@"}
	PARSER_RULE_CONTEXT_OPTIONAL_CHAINING_TOKEN               = ParserRuleContext{value: "?."}
	PARSER_RULE_CONTEXT_DOT_LT_TOKEN                          = ParserRuleContext{value: ".<"}
	PARSER_RULE_CONTEXT_SLASH_LT_TOKEN                        = ParserRuleContext{value: "/<"}
	PARSER_RULE_CONTEXT_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN = ParserRuleContext{value: "/**/<"}
	PARSER_RULE_CONTEXT_SLASH_ASTERISK_TOKEN                  = ParserRuleContext{value: "/*"}
	PARSER_RULE_CONTEXT_RIGHT_DOUBLE_ARROW                    = ParserRuleContext{value: "=>"}
	PARSER_RULE_CONTEXT_DOUBLE_LT                             = ParserRuleContext{value: "<<"}
	PARSER_RULE_CONTEXT_DOUBLE_EQUAL                          = ParserRuleContext{value: "=="}
	PARSER_RULE_CONTEXT_BITWISE_XOR                           = ParserRuleContext{value: "^"}
	PARSER_RULE_CONTEXT_LOGICAL_AND                           = ParserRuleContext{value: "&&"}
	PARSER_RULE_CONTEXT_LOGICAL_OR                            = ParserRuleContext{value: "||"}
	PARSER_RULE_CONTEXT_ELVIS                                 = ParserRuleContext{value: "?:"}

	// Other terminals
	PARSER_RULE_CONTEXT_FUNC_NAME                        = ParserRuleContext{value: "func-name"}
	PARSER_RULE_CONTEXT_VARIABLE_NAME                    = ParserRuleContext{value: "variable-name"}
	PARSER_RULE_CONTEXT_SIMPLE_TYPE_DESCRIPTOR           = ParserRuleContext{value: "simple-type-desc"}
	PARSER_RULE_CONTEXT_BINARY_OPERATOR                  = ParserRuleContext{value: "binary-operator"}
	PARSER_RULE_CONTEXT_TYPE_NAME                        = ParserRuleContext{value: "type-name"}
	PARSER_RULE_CONTEXT_CLASS_NAME                       = ParserRuleContext{value: "class-name"}
	PARSER_RULE_CONTEXT_BOOLEAN_LITERAL                  = ParserRuleContext{value: "boolean-literal"}
	PARSER_RULE_CONTEXT_CHECKING_KEYWORD                 = ParserRuleContext{value: "checking-keyword"}
	PARSER_RULE_CONTEXT_COMPOUND_BINARY_OPERATOR         = ParserRuleContext{value: "compound-binary-operator"}
	PARSER_RULE_CONTEXT_UNARY_OPERATOR                   = ParserRuleContext{value: "unary-operator"}
	PARSER_RULE_CONTEXT_FUNCTION_IDENT                   = ParserRuleContext{value: "func-ident"}
	PARSER_RULE_CONTEXT_FIELD_IDENT                      = ParserRuleContext{value: "field-ident"}
	PARSER_RULE_CONTEXT_OBJECT_IDENT                     = ParserRuleContext{value: "object-ident"}
	PARSER_RULE_CONTEXT_SERVICE_IDENT                    = ParserRuleContext{value: "service-ident"}
	PARSER_RULE_CONTEXT_SERVICE_IDENT_RHS                = ParserRuleContext{value: "service-ident-rhs"}
	PARSER_RULE_CONTEXT_REMOTE_IDENT                     = ParserRuleContext{value: "remote-ident"}
	PARSER_RULE_CONTEXT_RECORD_IDENT                     = ParserRuleContext{value: "record-ident"}
	PARSER_RULE_CONTEXT_ANNOTATION_TAG                   = ParserRuleContext{value: "annotation-tag"}
	PARSER_RULE_CONTEXT_ATTACH_POINT_END                 = ParserRuleContext{value: "attach-point-end"}
	PARSER_RULE_CONTEXT_IDENTIFIER                       = ParserRuleContext{value: "identifier"}
	PARSER_RULE_CONTEXT_PATH_SEGMENT_IDENT               = ParserRuleContext{value: "path-segment-ident"}
	PARSER_RULE_CONTEXT_NAMESPACE_PREFIX                 = ParserRuleContext{value: "namespace-prefix"}
	PARSER_RULE_CONTEXT_WORKER_NAME                      = ParserRuleContext{value: "worker-name"}
	PARSER_RULE_CONTEXT_FIELD_OR_FUNC_NAME               = ParserRuleContext{value: "field-or-func-name"}
	PARSER_RULE_CONTEXT_ORDER_DIRECTION                  = ParserRuleContext{value: "order-direction"}
	PARSER_RULE_CONTEXT_VAR_REF_COLON                    = ParserRuleContext{value: "var-ref-colon"}
	PARSER_RULE_CONTEXT_TYPE_REF_COLON                   = ParserRuleContext{value: "type-ref-colon"}
	PARSER_RULE_CONTEXT_METHOD_CALL_DOT                  = ParserRuleContext{value: "method-call-dot"}
	PARSER_RULE_CONTEXT_RESOURCE_METHOD_CALL_SLASH_TOKEN = ParserRuleContext{value: "resource-method-call-slash-token"}

	// Expressions
	PARSER_RULE_CONTEXT_EXPRESSION                                             = ParserRuleContext{value: "expression"}
	PARSER_RULE_CONTEXT_TERMINAL_EXPRESSION                                    = ParserRuleContext{value: "terminal-expression"}
	PARSER_RULE_CONTEXT_EXPRESSION_RHS                                         = ParserRuleContext{value: "expression-rhs"}
	PARSER_RULE_CONTEXT_FUNC_CALL                                              = ParserRuleContext{value: "func-call"}
	PARSER_RULE_CONTEXT_BASIC_LITERAL                                          = ParserRuleContext{value: "basic-literal"}
	PARSER_RULE_CONTEXT_ACCESS_EXPRESSION                                      = ParserRuleContext{value: "access-expr"} // method-call, field-access, member-access
	PARSER_RULE_CONTEXT_DECIMAL_INTEGER_LITERAL_TOKEN                          = ParserRuleContext{value: "decimal-int-literal-token"}
	PARSER_RULE_CONTEXT_VARIABLE_REF                                           = ParserRuleContext{value: "var-ref"}
	PARSER_RULE_CONTEXT_STRING_LITERAL_TOKEN                                   = ParserRuleContext{value: "string-literal-token"}
	PARSER_RULE_CONTEXT_MAPPING_CONSTRUCTOR                                    = ParserRuleContext{value: "mapping-constructor"}
	PARSER_RULE_CONTEXT_MAPPING_FIELD                                          = ParserRuleContext{value: "maping-field"}
	PARSER_RULE_CONTEXT_FIRST_MAPPING_FIELD                                    = ParserRuleContext{value: "first-mapping-field"}
	PARSER_RULE_CONTEXT_MAPPING_FIELD_NAME                                     = ParserRuleContext{value: "maping-field-name"}
	PARSER_RULE_CONTEXT_SPECIFIC_FIELD_RHS                                     = ParserRuleContext{value: "specific-field-rhs"}
	PARSER_RULE_CONTEXT_SPECIFIC_FIELD                                         = ParserRuleContext{value: "specific-field"}
	PARSER_RULE_CONTEXT_COMPUTED_FIELD_NAME                                    = ParserRuleContext{value: "computed-field-name"}
	PARSER_RULE_CONTEXT_MAPPING_FIELD_END                                      = ParserRuleContext{value: "mapping-field-end"}
	PARSER_RULE_CONTEXT_TYPEOF_EXPRESSION                                      = ParserRuleContext{value: "typeof-expr"}
	PARSER_RULE_CONTEXT_UNARY_EXPRESSION                                       = ParserRuleContext{value: "unary-expr"}
	PARSER_RULE_CONTEXT_HEX_INTEGER_LITERAL_TOKEN                              = ParserRuleContext{value: "hex-integer-literal-token"}
	PARSER_RULE_CONTEXT_NIL_LITERAL                                            = ParserRuleContext{value: "nil-literal"}
	PARSER_RULE_CONTEXT_CONSTANT_EXPRESSION                                    = ParserRuleContext{value: "constant-expr"}
	PARSER_RULE_CONTEXT_CONSTANT_EXPRESSION_START                              = ParserRuleContext{value: "constant-expr-start"}
	PARSER_RULE_CONTEXT_DECIMAL_FLOATING_POINT_LITERAL_TOKEN                   = ParserRuleContext{value: "decimal-floating-point-literal-token"}
	PARSER_RULE_CONTEXT_HEX_FLOATING_POINT_LITERAL_TOKEN                       = ParserRuleContext{value: "hex-floating-point-literal-token"}
	PARSER_RULE_CONTEXT_LIST_CONSTRUCTOR                                       = ParserRuleContext{value: "list-constructor"}
	PARSER_RULE_CONTEXT_LIST_CONSTRUCTOR_FIRST_MEMBER                          = ParserRuleContext{value: "list-constructor-first-member"}
	PARSER_RULE_CONTEXT_LIST_CONSTRUCTOR_MEMBER                                = ParserRuleContext{value: "list-constructor-member"}
	PARSER_RULE_CONTEXT_TYPE_CAST                                              = ParserRuleContext{value: "type-cast"}
	PARSER_RULE_CONTEXT_TYPE_CAST_PARAM                                        = ParserRuleContext{value: "type-cast-param"}
	PARSER_RULE_CONTEXT_TYPE_CAST_PARAM_RHS                                    = ParserRuleContext{value: "type-cast-param-rhs"}
	PARSER_RULE_CONTEXT_TYPE_CAST_PARAM_START                                  = ParserRuleContext{value: "type-cast-param-start"}
	PARSER_RULE_CONTEXT_TABLE_CONSTRUCTOR                                      = ParserRuleContext{value: "table-constructor"}
	PARSER_RULE_CONTEXT_TABLE_KEYWORD_RHS                                      = ParserRuleContext{value: "table-keyword-rhs"}
	PARSER_RULE_CONTEXT_ROW_LIST_RHS                                           = ParserRuleContext{value: "row-list-rhs"}
	PARSER_RULE_CONTEXT_TABLE_ROW_END                                          = ParserRuleContext{value: "table-row-end"}
	PARSER_RULE_CONTEXT_NEW_KEYWORD_RHS                                        = ParserRuleContext{value: "new-keyword-rhs"}
	PARSER_RULE_CONTEXT_IMPLICIT_NEW                                           = ParserRuleContext{value: "implicit-new"}
	PARSER_RULE_CONTEXT_CLASS_DESCRIPTOR_IN_NEW_EXPR                           = ParserRuleContext{value: "class-descriptor-in-new-expr"}
	PARSER_RULE_CONTEXT_LET_EXPRESSION                                         = ParserRuleContext{value: "let-expr"}
	PARSER_RULE_CONTEXT_ANON_FUNC_EXPRESSION                                   = ParserRuleContext{value: "anon-func-expression"}
	PARSER_RULE_CONTEXT_ANON_FUNC_EXPRESSION_START                             = ParserRuleContext{value: "anon-func-expression-start"}
	PARSER_RULE_CONTEXT_TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION                  = ParserRuleContext{value: "table-constructor-or-query-expr"}
	PARSER_RULE_CONTEXT_TABLE_CONSTRUCTOR_OR_QUERY_START                       = ParserRuleContext{value: "table-constructor-or-query-start"}
	PARSER_RULE_CONTEXT_TABLE_CONSTRUCTOR_OR_QUERY_RHS                         = ParserRuleContext{value: "table-constructor-or-query-rhs"}
	PARSER_RULE_CONTEXT_QUERY_EXPRESSION                                       = ParserRuleContext{value: "query-expr"}
	PARSER_RULE_CONTEXT_QUERY_EXPRESSION_RHS                                   = ParserRuleContext{value: "query-expr-rhs"}
	PARSER_RULE_CONTEXT_QUERY_ACTION_RHS                                       = ParserRuleContext{value: "query-action-rhs"}
	PARSER_RULE_CONTEXT_QUERY_EXPRESSION_END                                   = ParserRuleContext{value: "query-expr-end"}
	PARSER_RULE_CONTEXT_FIELD_ACCESS_IDENTIFIER                                = ParserRuleContext{value: "field-access-identifier"}
	PARSER_RULE_CONTEXT_QUERY_PIPELINE_RHS                                     = ParserRuleContext{value: "query-pipeline-rhs"}
	PARSER_RULE_CONTEXT_LET_CLAUSE_END                                         = ParserRuleContext{value: "let-clause-end"}
	PARSER_RULE_CONTEXT_CONDITIONAL_EXPRESSION                                 = ParserRuleContext{value: "conditional-expr"}
	PARSER_RULE_CONTEXT_XML_NAVIGATE_EXPR                                      = ParserRuleContext{value: "xml-navigate-expr"}
	PARSER_RULE_CONTEXT_XML_FILTER_EXPR                                        = ParserRuleContext{value: "xml-filter-expr"}
	PARSER_RULE_CONTEXT_XML_STEP_EXPR                                          = ParserRuleContext{value: "xml-step-expr"}
	PARSER_RULE_CONTEXT_XML_NAME_PATTERN                                       = ParserRuleContext{value: "xml-name-pattern"}
	PARSER_RULE_CONTEXT_XML_NAME_PATTERN_RHS                                   = ParserRuleContext{value: "xml-name-pattern-rhs"}
	PARSER_RULE_CONTEXT_XML_ATOMIC_NAME_PATTERN                                = ParserRuleContext{value: "xml-atomic_name-pattern"}
	PARSER_RULE_CONTEXT_XML_ATOMIC_NAME_PATTERN_START                          = ParserRuleContext{value: "xml-atomic_name-pattern-start"}
	PARSER_RULE_CONTEXT_XML_ATOMIC_NAME_IDENTIFIER                             = ParserRuleContext{value: "xml-atomic_name-identifier"}
	PARSER_RULE_CONTEXT_XML_ATOMIC_NAME_IDENTIFIER_RHS                         = ParserRuleContext{value: "xml-atomic_name-identifier-rhs"}
	PARSER_RULE_CONTEXT_XML_STEP_START                                         = ParserRuleContext{value: "xml-step-start"}
	PARSER_RULE_CONTEXT_VARIABLE_REF_RHS                                       = ParserRuleContext{value: "variable-ref-rhs"}
	PARSER_RULE_CONTEXT_ORDER_CLAUSE_END                                       = ParserRuleContext{value: "order-clause-end"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR                                     = ParserRuleContext{value: "object-constructor"}
	PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_TYPE_REF                            = ParserRuleContext{value: "object-constructor-type-ref"}
	PARSER_RULE_CONTEXT_ERROR_CONSTRUCTOR                                      = ParserRuleContext{value: "error-constructor"}
	PARSER_RULE_CONTEXT_ERROR_CONSTRUCTOR_RHS                                  = ParserRuleContext{value: "error-constructor-rhs"}
	PARSER_RULE_CONTEXT_INFERRED_TYPEDESC_DEFAULT_START_LT                     = ParserRuleContext{value: "inferred-typedesc-default-start-lt"}
	PARSER_RULE_CONTEXT_INFERRED_TYPEDESC_DEFAULT_END_GT                       = ParserRuleContext{value: "inferred-typedesc-default-end-gt"}
	PARSER_RULE_CONTEXT_EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START          = ParserRuleContext{value: "expr-start-or-inferred-typedesc-default-start"}
	PARSER_RULE_CONTEXT_TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END = ParserRuleContext{value: "type-cast-param-start-or-inferred-typedesc-default-end"}
	PARSER_RULE_CONTEXT_END_OF_PARAMS_OR_NEXT_PARAM_START                      = ParserRuleContext{value: "end-of-params-or-next-param-start"}
	PARSER_RULE_CONTEXT_BRACED_EXPRESSION                                      = ParserRuleContext{value: "braced-expression"}
	PARSER_RULE_CONTEXT_ACTION                                                 = ParserRuleContext{value: "action"}

	// Contexts that expect a type
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_ANNOTATION_DECL                = ParserRuleContext{value: "type-desc-annotation-descl"}
	PARSER_RULE_CONTEXT_TYPE_DESC_BEFORE_IDENTIFIER                 = ParserRuleContext{value: "type-desc-before-identifier"} // object/record fields, params, const, listener
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_RECORD_FIELD                   = ParserRuleContext{value: "type-desc-in-record-field"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_PARAM                          = ParserRuleContext{value: "type-desc-in-param"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_TYPE_BINDING_PATTERN           = ParserRuleContext{value: "type-desc-in-type-binding-pattern"} // foreach, let-var-decl, var-decl
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_TYPE_DEF                       = ParserRuleContext{value: "type-def-type-desc"}                // local/mdule type defitions
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_ANGLE_BRACKETS                 = ParserRuleContext{value: "type-desc-in-angle-bracket"}        // type-cast, parameterized-type
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_RETURN_TYPE_DESC               = ParserRuleContext{value: "type-desc-in-return-type-desc"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_EXPRESSION                     = ParserRuleContext{value: "type-desc-in-expression"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_STREAM_TYPE_DESC               = ParserRuleContext{value: "type-desc-in-stream-type-desc"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_TUPLE                          = ParserRuleContext{value: "type-desc-in-tuple"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_PARENTHESIS                    = ParserRuleContext{value: "type-desc-in-parenthesis"}
	PARSER_RULE_CONTEXT_VAR_DECL_STARTED_WITH_DENTIFIER             = ParserRuleContext{value: "var-decl-started-with-dentifier"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_SERVICE                        = ParserRuleContext{value: "type-desc-in-service"}
	PARSER_RULE_CONTEXT_TYPE_DESC_IN_PATH_PARAM                     = ParserRuleContext{value: "type-desc-in-path-param"}
	PARSER_RULE_CONTEXT_TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY = ParserRuleContext{value: "type-desc-before-identifier-in-grouping-key"}

	// XML
	PARSER_RULE_CONTEXT_XML_CONTENT                = ParserRuleContext{value: "xml-content"}
	PARSER_RULE_CONTEXT_XML_TAG                    = ParserRuleContext{value: "xml-tag"}
	PARSER_RULE_CONTEXT_XML_START_OR_EMPTY_TAG     = ParserRuleContext{value: "xml-start-or-empty-tag"}
	PARSER_RULE_CONTEXT_XML_START_OR_EMPTY_TAG_END = ParserRuleContext{value: "xml-start-or-empty-tag-end"}
	PARSER_RULE_CONTEXT_XML_END_TAG                = ParserRuleContext{value: "xml-end-tag"}
	PARSER_RULE_CONTEXT_XML_NAME                   = ParserRuleContext{value: "xml-name"}
	PARSER_RULE_CONTEXT_XML_PI                     = ParserRuleContext{value: "xml-pi"}
	PARSER_RULE_CONTEXT_XML_TEXT                   = ParserRuleContext{value: "xml-text"}
	PARSER_RULE_CONTEXT_XML_ATTRIBUTES             = ParserRuleContext{value: "xml-attributes"}
	PARSER_RULE_CONTEXT_XML_ATTRIBUTE              = ParserRuleContext{value: "xml-attribute"}
	PARSER_RULE_CONTEXT_XML_ATTRIBUTE_VALUE_ITEM   = ParserRuleContext{value: "xml-attribute-value-item"}
	PARSER_RULE_CONTEXT_XML_ATTRIBUTE_VALUE_TEXT   = ParserRuleContext{value: "xml-attribute-value-text"}
	PARSER_RULE_CONTEXT_XML_COMMENT_START          = ParserRuleContext{value: "<!--"}
	PARSER_RULE_CONTEXT_XML_COMMENT_END            = ParserRuleContext{value: "-->"}
	PARSER_RULE_CONTEXT_XML_COMMENT_CONTENT        = ParserRuleContext{value: "xml-comment-content"}
	PARSER_RULE_CONTEXT_XML_PI_START               = ParserRuleContext{value: "<?"}
	PARSER_RULE_CONTEXT_XML_PI_END                 = ParserRuleContext{value: "?>"}
	PARSER_RULE_CONTEXT_XML_PI_DATA                = ParserRuleContext{value: "xml-pi-data"}
	PARSER_RULE_CONTEXT_XML_PI_TARGET_RHS          = ParserRuleContext{value: "xml-pi-target-rhs"}
	PARSER_RULE_CONTEXT_INTERPOLATION_START_TOKEN  = ParserRuleContext{value: "${"}
	PARSER_RULE_CONTEXT_INTERPOLATION              = ParserRuleContext{value: "interoplation"}
	PARSER_RULE_CONTEXT_TEMPLATE_BODY              = ParserRuleContext{value: "template-body"}
	PARSER_RULE_CONTEXT_TEMPLATE_MEMBER            = ParserRuleContext{value: "template-member"}
	PARSER_RULE_CONTEXT_TEMPLATE_STRING            = ParserRuleContext{value: "template-string"}
	PARSER_RULE_CONTEXT_TEMPLATE_STRING_RHS        = ParserRuleContext{value: "template-string-rhs"}
	PARSER_RULE_CONTEXT_XML_QUOTE_START            = ParserRuleContext{value: "xml-quote-start"}
	PARSER_RULE_CONTEXT_XML_QUOTE_END              = ParserRuleContext{value: "xml-quote-end"}
	PARSER_RULE_CONTEXT_XML_CDATA_START            = ParserRuleContext{value: "xml-cdata-start"}
	PARSER_RULE_CONTEXT_XML_OPTIONAL_CDATA_CONTENT = ParserRuleContext{value: "xml-optional-cdata-content"}
	PARSER_RULE_CONTEXT_XML_CDATA_CONTENT          = ParserRuleContext{value: "xml-cdata-content"}
	PARSER_RULE_CONTEXT_XML_CDATA_END              = ParserRuleContext{value: "xml-cdata-end"}

	//Other
	PARSER_RULE_CONTEXT_TYPE_DESC_RHS                            = ParserRuleContext{value: "type-desc-rhs"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_FUNC_KEYWORD_RHS               = ParserRuleContext{value: "func-type-func-keyword-rhs"}
	PARSER_RULE_CONTEXT_FUNC_TYPE_FUNC_KEYWORD_RHS_START         = ParserRuleContext{value: "func-type-func-keyword-rhs-start"}
	PARSER_RULE_CONTEXT_STREAM_TYPE_PARAM_START_TOKEN            = ParserRuleContext{value: "stream-type-param-start-token"}
	PARSER_RULE_CONTEXT_STREAM_TYPE_FIRST_PARAM_RHS              = ParserRuleContext{value: "stream-type-params"}
	PARSER_RULE_CONTEXT_KEY_CONSTRAINTS_RHS                      = ParserRuleContext{value: "key-constraints-rhs"}
	PARSER_RULE_CONTEXT_ROW_TYPE_PARAM                           = ParserRuleContext{value: "row-type-param"}
	PARSER_RULE_CONTEXT_TABLE_TYPE_DESC_RHS                      = ParserRuleContext{value: "table-type-desc-rhs"}
	PARSER_RULE_CONTEXT_SIGNED_INT_OR_FLOAT_RHS                  = ParserRuleContext{value: "signed-int-or-float-rhs"}
	PARSER_RULE_CONTEXT_ENUM_MEMBER_LIST                         = ParserRuleContext{value: "enum-member-list"}
	PARSER_RULE_CONTEXT_ENUM_MEMBER_END                          = ParserRuleContext{value: "enum-member-rhs"}
	PARSER_RULE_CONTEXT_ENUM_MEMBER_RHS                          = ParserRuleContext{value: "enum-member-internal-rhs"}
	PARSER_RULE_CONTEXT_ENUM_MEMBER_START                        = ParserRuleContext{value: "enum-member-start"}
	PARSER_RULE_CONTEXT_TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER     = ParserRuleContext{value: "tuple-type-desc-or-list-cont-member"}
	PARSER_RULE_CONTEXT_MAP_TYPE_OR_TYPE_REF                     = ParserRuleContext{value: "map-type-or-type-ref"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_OR_TYPE_REF                  = ParserRuleContext{value: "object-type-or-type-ref"}
	PARSER_RULE_CONTEXT_STREAM_TYPE_OR_TYPE_REF                  = ParserRuleContext{value: "stream-type-or-type-ref"}
	PARSER_RULE_CONTEXT_TABLE_TYPE_OR_TYPE_REF                   = ParserRuleContext{value: "table-type-or-type-ref"}
	PARSER_RULE_CONTEXT_PARAMETERIZED_TYPE_OR_TYPE_REF           = ParserRuleContext{value: "parameterized-type-or-type-ref"}
	PARSER_RULE_CONTEXT_TYPE_DESC_RHS_OR_TYPE_REF                = ParserRuleContext{value: "type-desc-rhs-or-type-ref"}
	PARSER_RULE_CONTEXT_OBJECT_TYPE_OBJECT_KEYWORD_RHS           = ParserRuleContext{value: "object-type-object-keyword-rhs"}
	PARSER_RULE_CONTEXT_TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF      = ParserRuleContext{value: "table-cons-or-query-expr-or-var-ref"}
	PARSER_RULE_CONTEXT_EXPRESSION_START_TABLE_KEYWORD_RHS       = ParserRuleContext{value: "expression-start-table-keyword-rhs"}
	PARSER_RULE_CONTEXT_QUERY_EXPR_OR_VAR_REF                    = ParserRuleContext{value: "query-expr-or-var-ref"}
	PARSER_RULE_CONTEXT_QUERY_CONSTRUCT_TYPE_RHS                 = ParserRuleContext{value: "query-construct-type-rhs"}
	PARSER_RULE_CONTEXT_ERROR_CONS_EXPR_OR_VAR_REF               = ParserRuleContext{value: "error-cons-expr-or-var-ref"}
	PARSER_RULE_CONTEXT_ERROR_CONS_ERROR_KEYWORD_RHS             = ParserRuleContext{value: "error-cons-error-keyword-rhs"}
	PARSER_RULE_CONTEXT_TRANSACTION_STMT_TRANSACTION_KEYWORD_RHS = ParserRuleContext{value: "transaction-stmt-transaction-keyword-rhs"}
	PARSER_RULE_CONTEXT_TRANSACTION_STMT_RHS_OR_TYPE_REF         = ParserRuleContext{value: "transaction-stmt-rhs-or-type-ref"}
	PARSER_RULE_CONTEXT_QUALIFIED_IDENTIFIER_START_IDENTIFIER    = ParserRuleContext{value: "qualified-identifier-start-identifier"}
	PARSER_RULE_CONTEXT_QUALIFIED_IDENTIFIER_PREDECLARED_PREFIX  = ParserRuleContext{value: "qualified-identifier-predeclared-prefix"}
	PARSER_RULE_CONTEXT_TYPE_DESC_RHS_OR_BP_RHS                  = ParserRuleContext{value: "type-desc-rhs-or-binding-pattern-rhs"}
	PARSER_RULE_CONTEXT_LIST_BINDING_PATTERN_RHS                 = ParserRuleContext{value: "list-binding-pattern-rhs"}
	PARSER_RULE_CONTEXT_TYPE_DESC_RHS_IN_TYPED_BP                = ParserRuleContext{value: "type-desc-rhs-in-typed-binding-pattern"}
	PARSER_RULE_CONTEXT_ASSIGNMENT_STMT_RHS                      = ParserRuleContext{value: "assignment-stmt-rhs"}
	PARSER_RULE_CONTEXT_ANNOTATION_DECL_START                    = ParserRuleContext{value: "annotation-declaration-start"}
	PARSER_RULE_CONTEXT_OPTIONAL_TOP_LEVEL_SEMICOLON             = ParserRuleContext{value: "optional-top-level-semicolon"}
	PARSER_RULE_CONTEXT_TUPLE_MEMBERS                            = ParserRuleContext{value: "tuple-members"}
	PARSER_RULE_CONTEXT_TUPLE_MEMBER                             = ParserRuleContext{value: "tuple-member"}
	PARSER_RULE_CONTEXT_SINGLE_OR_ALTERNATE_WORKER               = ParserRuleContext{value: "single-or-alternate-worker"}
	PARSER_RULE_CONTEXT_SINGLE_OR_ALTERNATE_WORKER_SEPARATOR     = ParserRuleContext{value: "single-or-alternate-worker-separator"}
	PARSER_RULE_CONTEXT_SINGLE_OR_ALTERNATE_WORKER_END           = ParserRuleContext{value: "single-or-alternate-worker-end"}
	PARSER_RULE_CONTEXT_XML_STEP_EXTENDS                         = ParserRuleContext{value: "xml-step-extends"}
	PARSER_RULE_CONTEXT_XML_STEP_EXTEND                          = ParserRuleContext{value: "xml-step-extend"}
	PARSER_RULE_CONTEXT_XML_STEP_EXTEND_END                      = ParserRuleContext{value: "xml-step-extend-end"}
	PARSER_RULE_CONTEXT_XML_STEP_START_END                       = ParserRuleContext{value: "xml-step-start-end"}
)

func (p ParserRuleContext) GetErrorCode() diagnostics.DiagnosticCode {
	switch p {
	case PARSER_RULE_CONTEXT_EXTERNAL_FUNC_BODY:
		return &ERROR_MISSING_EQUAL_TOKEN
	case PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK:
		return &ERROR_MISSING_OPEN_BRACE_TOKEN
	case PARSER_RULE_CONTEXT_FUNC_DEF,
		PARSER_RULE_CONTEXT_FUNC_DEF_OR_FUNC_TYPE,
		PARSER_RULE_CONTEXT_FUNC_TYPE_DESC,
		PARSER_RULE_CONTEXT_FUNC_TYPE_DESC_OR_ANON_FUNC,
		PARSER_RULE_CONTEXT_IDENT_AFTER_OBJECT_IDENT,
		PARSER_RULE_CONTEXT_FUNC_DEF_FIRST_QUALIFIER,
		PARSER_RULE_CONTEXT_FUNC_DEF_SECOND_QUALIFIER,
		PARSER_RULE_CONTEXT_FUNC_TYPE_FIRST_QUALIFIER,
		PARSER_RULE_CONTEXT_FUNC_TYPE_SECOND_QUALIFIER,
		PARSER_RULE_CONTEXT_OBJECT_METHOD_FIRST_QUALIFIER,
		PARSER_RULE_CONTEXT_OBJECT_METHOD_SECOND_QUALIFIER,
		PARSER_RULE_CONTEXT_OBJECT_METHOD_THIRD_QUALIFIER,
		PARSER_RULE_CONTEXT_OBJECT_METHOD_FOURTH_QUALIFIER:
		return &ERROR_MISSING_FUNCTION_KEYWORD
	case PARSER_RULE_CONTEXT_SINGLE_KEYWORD_ATTACH_POINT_IDENT:
		return &ERROR_MISSING_ATTACH_POINT_NAME
	case PARSER_RULE_CONTEXT_SIMPLE_TYPE_DESCRIPTOR:
		return &ERROR_MISSING_BUILTIN_TYPE
	case PARSER_RULE_CONTEXT_REQUIRED_PARAM,
		PARSER_RULE_CONTEXT_VAR_DECL_STMT,
		PARSER_RULE_CONTEXT_ASSIGNMENT_OR_VAR_DECL_STMT,
		PARSER_RULE_CONTEXT_DEFAULTABLE_PARAM,
		PARSER_RULE_CONTEXT_REST_PARAM,
		PARSER_RULE_CONTEXT_TYPE_DESCRIPTOR,
		PARSER_RULE_CONTEXT_OPTIONAL_TYPE_DESCRIPTOR,
		PARSER_RULE_CONTEXT_ARRAY_TYPE_DESCRIPTOR,
		PARSER_RULE_CONTEXT_SIMPLE_TYPE_DESC_IDENTIFIER:
		return &ERROR_MISSING_TYPE_DESC
	case PARSER_RULE_CONTEXT_TYPE_REFERENCE:
		return &ERROR_MISSING_TYPE_REFERENCE
	case PARSER_RULE_CONTEXT_TYPE_NAME,
		PARSER_RULE_CONTEXT_TYPE_REFERENCE_IN_TYPE_INCLUSION,
		PARSER_RULE_CONTEXT_FIELD_ACCESS_IDENTIFIER,
		PARSER_RULE_CONTEXT_CLASS_NAME,
		PARSER_RULE_CONTEXT_FUNC_NAME,
		PARSER_RULE_CONTEXT_VARIABLE_NAME,
		PARSER_RULE_CONTEXT_IMPORT_MODULE_NAME,
		PARSER_RULE_CONTEXT_IMPORT_ORG_OR_MODULE_NAME,
		PARSER_RULE_CONTEXT_IMPORT_PREFIX,
		PARSER_RULE_CONTEXT_VARIABLE_REF,
		PARSER_RULE_CONTEXT_BASIC_LITERAL,
		PARSER_RULE_CONTEXT_IDENTIFIER,
		PARSER_RULE_CONTEXT_QUALIFIED_IDENTIFIER_START_IDENTIFIER,
		PARSER_RULE_CONTEXT_NAMESPACE_PREFIX,
		PARSER_RULE_CONTEXT_IMPLICIT_ANON_FUNC_PARAM,
		PARSER_RULE_CONTEXT_METHOD_NAME,
		PARSER_RULE_CONTEXT_PEER_WORKER_NAME,
		PARSER_RULE_CONTEXT_RECEIVE_FIELD_NAME,
		PARSER_RULE_CONTEXT_WAIT_FIELD_NAME,
		PARSER_RULE_CONTEXT_FIELD_BINDING_PATTERN_NAME,
		PARSER_RULE_CONTEXT_XML_ATOMIC_NAME_IDENTIFIER,
		PARSER_RULE_CONTEXT_MAPPING_FIELD_NAME,
		PARSER_RULE_CONTEXT_WORKER_NAME,
		PARSER_RULE_CONTEXT_NAMED_WORKERS,
		PARSER_RULE_CONTEXT_ANNOTATION_TAG,
		PARSER_RULE_CONTEXT_AFTER_PARAMETER_TYPE,
		PARSER_RULE_CONTEXT_MODULE_ENUM_NAME,
		PARSER_RULE_CONTEXT_ENUM_MEMBER_NAME,
		PARSER_RULE_CONTEXT_TYPED_BINDING_PATTERN_TYPE_RHS,
		PARSER_RULE_CONTEXT_ASSIGNMENT_STMT,
		PARSER_RULE_CONTEXT_XML_NAME,
		PARSER_RULE_CONTEXT_ACCESS_EXPRESSION,
		PARSER_RULE_CONTEXT_BINDING_PATTERN_STARTING_IDENTIFIER,
		PARSER_RULE_CONTEXT_COMPUTED_FIELD_NAME,
		PARSER_RULE_CONTEXT_SIMPLE_BINDING_PATTERN,
		PARSER_RULE_CONTEXT_ERROR_FIELD_BINDING_PATTERN,
		PARSER_RULE_CONTEXT_ERROR_CAUSE_SIMPLE_BINDING_PATTERN,
		PARSER_RULE_CONTEXT_PATH_SEGMENT_IDENT,
		PARSER_RULE_CONTEXT_NAMED_ARG_BINDING_PATTERN,
		PARSER_RULE_CONTEXT_MODULE_VAR_FIRST_QUAL,
		PARSER_RULE_CONTEXT_MODULE_VAR_SECOND_QUAL,
		PARSER_RULE_CONTEXT_MODULE_VAR_THIRD_QUAL,
		PARSER_RULE_CONTEXT_OBJECT_MEMBER_VISIBILITY_QUAL:
		return &ERROR_MISSING_IDENTIFIER
	case PARSER_RULE_CONTEXT_EXPRESSION,
		PARSER_RULE_CONTEXT_TERMINAL_EXPRESSION:
		return &ERROR_MISSING_EXPRESSION
	case PARSER_RULE_CONTEXT_STRING_LITERAL_TOKEN:
		return &ERROR_MISSING_STRING_LITERAL
	case PARSER_RULE_CONTEXT_DECIMAL_INTEGER_LITERAL_TOKEN,
		PARSER_RULE_CONTEXT_SIGNED_INT_OR_FLOAT_RHS:
		return &ERROR_MISSING_DECIMAL_INTEGER_LITERAL
	case PARSER_RULE_CONTEXT_HEX_INTEGER_LITERAL_TOKEN:
		return &ERROR_MISSING_HEX_INTEGER_LITERAL
	case PARSER_RULE_CONTEXT_OBJECT_FIELD_RHS,
		PARSER_RULE_CONTEXT_BINDING_PATTERN_OR_VAR_REF_RHS:
		return &ERROR_MISSING_SEMICOLON_TOKEN
	case PARSER_RULE_CONTEXT_NIL_LITERAL,
		PARSER_RULE_CONTEXT_ERROR_MATCH_PATTERN:
		return &ERROR_MISSING_ERROR_KEYWORD
	case PARSER_RULE_CONTEXT_DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
		return &ERROR_MISSING_DECIMAL_FLOATING_POINT_LITERAL
	case PARSER_RULE_CONTEXT_HEX_FLOATING_POINT_LITERAL_TOKEN:
		return &ERROR_MISSING_HEX_FLOATING_POINT_LITERAL
	case PARSER_RULE_CONTEXT_STATEMENT,
		PARSER_RULE_CONTEXT_STATEMENT_WITHOUT_ANNOTS:
		return &ERROR_MISSING_CLOSE_BRACE_TOKEN
	case PARSER_RULE_CONTEXT_XML_COMMENT_CONTENT,
		PARSER_RULE_CONTEXT_XML_PI_DATA:
		return &ERROR_MISSING_XML_TEXT_CONTENT
	default:
		return p.getSeperatorTokenErrorCode()
	}
}

func (p ParserRuleContext) getSeperatorTokenErrorCode() diagnostics.DiagnosticCode {
	switch p {
	case PARSER_RULE_CONTEXT_BITWISE_AND_OPERATOR:
		return &ERROR_MISSING_BITWISE_AND_TOKEN
	case PARSER_RULE_CONTEXT_EQUAL_OR_RIGHT_ARROW,
		PARSER_RULE_CONTEXT_ASSIGN_OP:
		return &ERROR_MISSING_EQUAL_TOKEN
	case PARSER_RULE_CONTEXT_BINARY_OPERATOR,
		PARSER_RULE_CONTEXT_UNARY_OPERATOR,
		PARSER_RULE_CONTEXT_COMPOUND_BINARY_OPERATOR,
		PARSER_RULE_CONTEXT_UNARY_EXPRESSION,
		PARSER_RULE_CONTEXT_EXPRESSION_RHS,
		PARSER_RULE_CONTEXT_PLUS_TOKEN:
		return &ERROR_MISSING_BINARY_OPERATOR
	case PARSER_RULE_CONTEXT_CLOSE_BRACE:
		return &ERROR_MISSING_CLOSE_BRACE_TOKEN
	case PARSER_RULE_CONTEXT_CLOSE_PARENTHESIS,
		PARSER_RULE_CONTEXT_ARG_LIST_CLOSE_PAREN:
		return &ERROR_MISSING_CLOSE_PAREN_TOKEN
	case PARSER_RULE_CONTEXT_COMMA,
		PARSER_RULE_CONTEXT_ERROR_MESSAGE_BINDING_PATTERN_END_COMMA,
		PARSER_RULE_CONTEXT_ERROR_MESSAGE_MATCH_PATTERN_END_COMMA:
		return &ERROR_MISSING_COMMA_TOKEN
	case PARSER_RULE_CONTEXT_OPEN_BRACE:
		return &ERROR_MISSING_OPEN_BRACE_TOKEN
	case PARSER_RULE_CONTEXT_OPEN_PARENTHESIS,
		PARSER_RULE_CONTEXT_ARG_LIST_OPEN_PAREN,
		PARSER_RULE_CONTEXT_PARENTHESISED_TYPE_DESC_START:
		return &ERROR_MISSING_OPEN_PAREN_TOKEN
	case PARSER_RULE_CONTEXT_SEMICOLON,
		PARSER_RULE_CONTEXT_OBJECT_FIELD_RHS:
		return &ERROR_MISSING_SEMICOLON_TOKEN
	case PARSER_RULE_CONTEXT_ASTERISK:
		return &ERROR_MISSING_ASTERISK_TOKEN
	case PARSER_RULE_CONTEXT_CLOSED_RECORD_BODY_END:
		return &ERROR_MISSING_CLOSE_BRACE_PIPE_TOKEN
	case PARSER_RULE_CONTEXT_CLOSED_RECORD_BODY_START:
		return &ERROR_MISSING_OPEN_BRACE_PIPE_TOKEN
	case PARSER_RULE_CONTEXT_ELLIPSIS:
		return &ERROR_MISSING_ELLIPSIS_TOKEN
	case PARSER_RULE_CONTEXT_QUESTION_MARK:
		return &ERROR_MISSING_QUESTION_MARK_TOKEN
	case PARSER_RULE_CONTEXT_CLOSE_BRACKET:
		return &ERROR_MISSING_CLOSE_BRACKET_TOKEN
	case PARSER_RULE_CONTEXT_DOT,
		PARSER_RULE_CONTEXT_METHOD_CALL_DOT:
		return &ERROR_MISSING_DOT_TOKEN
	case PARSER_RULE_CONTEXT_OPEN_BRACKET,
		PARSER_RULE_CONTEXT_TUPLE_TYPE_DESC_START:
		return &ERROR_MISSING_OPEN_BRACKET_TOKEN
	case PARSER_RULE_CONTEXT_SLASH,
		PARSER_RULE_CONTEXT_ABSOLUTE_PATH_SINGLE_SLASH,
		PARSER_RULE_CONTEXT_RESOURCE_METHOD_CALL_SLASH_TOKEN:
		return &ERROR_MISSING_SLASH_TOKEN
	case PARSER_RULE_CONTEXT_COLON,
		PARSER_RULE_CONTEXT_VAR_REF_COLON,
		PARSER_RULE_CONTEXT_TYPE_REF_COLON:
		return &ERROR_MISSING_COLON_TOKEN
	case PARSER_RULE_CONTEXT_AT:
		return &ERROR_MISSING_AT_TOKEN
	case PARSER_RULE_CONTEXT_RIGHT_ARROW:
		return &ERROR_MISSING_RIGHT_ARROW_TOKEN
	case PARSER_RULE_CONTEXT_GT,
		PARSER_RULE_CONTEXT_GT_TOKEN,
		PARSER_RULE_CONTEXT_XML_START_OR_EMPTY_TAG_END,
		PARSER_RULE_CONTEXT_XML_ATTRIBUTES,
		PARSER_RULE_CONTEXT_INFERRED_TYPEDESC_DEFAULT_END_GT:
		return &ERROR_MISSING_GT_TOKEN
	case PARSER_RULE_CONTEXT_LT,
		PARSER_RULE_CONTEXT_LT_TOKEN,
		PARSER_RULE_CONTEXT_XML_START_OR_EMPTY_TAG,
		PARSER_RULE_CONTEXT_XML_END_TAG,
		PARSER_RULE_CONTEXT_INFERRED_TYPEDESC_DEFAULT_START_LT,
		PARSER_RULE_CONTEXT_STREAM_TYPE_PARAM_START_TOKEN:
		return &ERROR_MISSING_LT_TOKEN
	case PARSER_RULE_CONTEXT_SYNC_SEND_TOKEN:
		return &ERROR_MISSING_SYNC_SEND_TOKEN
	case PARSER_RULE_CONTEXT_ANNOT_CHAINING_TOKEN:
		return &ERROR_MISSING_ANNOT_CHAINING_TOKEN
	case PARSER_RULE_CONTEXT_OPTIONAL_CHAINING_TOKEN:
		return &ERROR_MISSING_OPTIONAL_CHAINING_TOKEN
	case PARSER_RULE_CONTEXT_DOT_LT_TOKEN:
		return &ERROR_MISSING_DOT_LT_TOKEN
	case PARSER_RULE_CONTEXT_SLASH_LT_TOKEN:
		return &ERROR_MISSING_SLASH_LT_TOKEN
	case PARSER_RULE_CONTEXT_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN:
		return &ERROR_MISSING_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN
	case PARSER_RULE_CONTEXT_SLASH_ASTERISK_TOKEN:
		return &ERROR_MISSING_SLASH_ASTERISK_TOKEN
	case PARSER_RULE_CONTEXT_MINUS_TOKEN:
		return &ERROR_MISSING_MINUS_TOKEN
	case PARSER_RULE_CONTEXT_LEFT_ARROW_TOKEN:
		return &ERROR_MISSING_LEFT_ARROW_TOKEN
	case PARSER_RULE_CONTEXT_TEMPLATE_END,
		PARSER_RULE_CONTEXT_TEMPLATE_START,
		PARSER_RULE_CONTEXT_XML_CONTENT,
		PARSER_RULE_CONTEXT_XML_TEXT:
		return &ERROR_MISSING_BACKTICK_TOKEN
	case PARSER_RULE_CONTEXT_XML_COMMENT_START:
		return &ERROR_MISSING_XML_COMMENT_START_TOKEN
	case PARSER_RULE_CONTEXT_XML_COMMENT_END:
		return &ERROR_MISSING_XML_COMMENT_END_TOKEN
	case PARSER_RULE_CONTEXT_XML_PI,
		PARSER_RULE_CONTEXT_XML_PI_START:
		return &ERROR_MISSING_XML_PI_START_TOKEN
	case PARSER_RULE_CONTEXT_XML_PI_END:
		return &ERROR_MISSING_XML_PI_END_TOKEN
	case PARSER_RULE_CONTEXT_XML_QUOTE_END,
		PARSER_RULE_CONTEXT_XML_QUOTE_START:
		return &ERROR_MISSING_DOUBLE_QUOTE_TOKEN
	case PARSER_RULE_CONTEXT_INTERPOLATION_START_TOKEN:
		return &ERROR_MISSING_INTERPOLATION_START_TOKEN
	case PARSER_RULE_CONTEXT_EXPR_FUNC_BODY_START,
		PARSER_RULE_CONTEXT_RIGHT_DOUBLE_ARROW:
		return &ERROR_MISSING_RIGHT_DOUBLE_ARROW_TOKEN
	case PARSER_RULE_CONTEXT_XML_CDATA_END:
		return &ERROR_MISSING_XML_CDATA_END_TOKEN
	default:
		return p.getKeywordErrorCode()
	}
}

func (p ParserRuleContext) getKeywordErrorCode() diagnostics.DiagnosticCode {
	switch p {
	case PARSER_RULE_CONTEXT_PUBLIC_KEYWORD:
		return &ERROR_MISSING_PUBLIC_KEYWORD
	case PARSER_RULE_CONTEXT_PRIVATE_KEYWORD:
		return &ERROR_MISSING_PRIVATE_KEYWORD
	case PARSER_RULE_CONTEXT_ABSTRACT_KEYWORD:
		return &ERROR_MISSING_ABSTRACT_KEYWORD
	case PARSER_RULE_CONTEXT_CLIENT_KEYWORD:
		return &ERROR_MISSING_CLIENT_KEYWORD
	case PARSER_RULE_CONTEXT_IMPORT_KEYWORD:
		return &ERROR_MISSING_IMPORT_KEYWORD
	case PARSER_RULE_CONTEXT_FUNCTION_KEYWORD,
		PARSER_RULE_CONTEXT_FUNCTION_IDENT,
		PARSER_RULE_CONTEXT_OPTIONAL_PEER_WORKER,
		PARSER_RULE_CONTEXT_DEFAULT_WORKER_NAME_IN_ASYNC_SEND:
		return &ERROR_MISSING_FUNCTION_KEYWORD
	case PARSER_RULE_CONTEXT_CONST_KEYWORD:
		return &ERROR_MISSING_CONST_KEYWORD
	case PARSER_RULE_CONTEXT_LISTENER_KEYWORD:
		return &ERROR_MISSING_LISTENER_KEYWORD
	case PARSER_RULE_CONTEXT_SERVICE_KEYWORD,
		PARSER_RULE_CONTEXT_SERVICE_IDENT,
		PARSER_RULE_CONTEXT_SERVICE_DECL_QUALIFIER:
		return &ERROR_MISSING_SERVICE_KEYWORD
	case PARSER_RULE_CONTEXT_XMLNS_KEYWORD,
		PARSER_RULE_CONTEXT_XML_NAMESPACE_DECLARATION:
		return &ERROR_MISSING_XMLNS_KEYWORD
	case PARSER_RULE_CONTEXT_ANNOTATION_KEYWORD:
		return &ERROR_MISSING_ANNOTATION_KEYWORD
	case PARSER_RULE_CONTEXT_TYPE_KEYWORD:
		return &ERROR_MISSING_TYPE_KEYWORD
	case PARSER_RULE_CONTEXT_RECORD_KEYWORD,
		PARSER_RULE_CONTEXT_RECORD_FIELD,
		PARSER_RULE_CONTEXT_RECORD_IDENT:
		return &ERROR_MISSING_RECORD_KEYWORD
	case PARSER_RULE_CONTEXT_OBJECT_KEYWORD,
		PARSER_RULE_CONTEXT_OBJECT_IDENT,
		PARSER_RULE_CONTEXT_OBJECT_TYPE_DESCRIPTOR,
		PARSER_RULE_CONTEXT_FIRST_OBJECT_CONS_QUALIFIER,
		PARSER_RULE_CONTEXT_SECOND_OBJECT_CONS_QUALIFIER,
		PARSER_RULE_CONTEXT_FIRST_OBJECT_TYPE_QUALIFIER,
		PARSER_RULE_CONTEXT_SECOND_OBJECT_TYPE_QUALIFIER:
		return &ERROR_MISSING_OBJECT_KEYWORD
	case PARSER_RULE_CONTEXT_AS_KEYWORD:
		return &ERROR_MISSING_AS_KEYWORD
	case PARSER_RULE_CONTEXT_ON_KEYWORD:
		return &ERROR_MISSING_ON_KEYWORD
	case PARSER_RULE_CONTEXT_FINAL_KEYWORD:
		return &ERROR_MISSING_FINAL_KEYWORD
	case PARSER_RULE_CONTEXT_SOURCE_KEYWORD:
		return &ERROR_MISSING_SOURCE_KEYWORD
	case PARSER_RULE_CONTEXT_WORKER_KEYWORD:
		return &ERROR_MISSING_WORKER_KEYWORD
	case PARSER_RULE_CONTEXT_FIELD_IDENT:
		return &ERROR_MISSING_FIELD_KEYWORD
	case PARSER_RULE_CONTEXT_RETURNS_KEYWORD:
		return &ERROR_MISSING_RETURNS_KEYWORD
	case PARSER_RULE_CONTEXT_RETURN_KEYWORD:
		return &ERROR_MISSING_RETURN_KEYWORD
	case PARSER_RULE_CONTEXT_EXTERNAL_KEYWORD:
		return &ERROR_MISSING_EXTERNAL_KEYWORD
	case PARSER_RULE_CONTEXT_BOOLEAN_LITERAL:
		return &ERROR_MISSING_TRUE_KEYWORD
	case PARSER_RULE_CONTEXT_IF_KEYWORD:
		return &ERROR_MISSING_IF_KEYWORD
	case PARSER_RULE_CONTEXT_ELSE_KEYWORD:
		return &ERROR_MISSING_ELSE_KEYWORD
	case PARSER_RULE_CONTEXT_WHILE_KEYWORD:
		return &ERROR_MISSING_WHILE_KEYWORD
	case PARSER_RULE_CONTEXT_CHECKING_KEYWORD:
		return &ERROR_MISSING_CHECK_KEYWORD
	case PARSER_RULE_CONTEXT_PANIC_KEYWORD:
		return &ERROR_MISSING_PANIC_KEYWORD
	case PARSER_RULE_CONTEXT_CONTINUE_KEYWORD:
		return &ERROR_MISSING_CONTINUE_KEYWORD
	case PARSER_RULE_CONTEXT_BREAK_KEYWORD:
		return &ERROR_MISSING_BREAK_KEYWORD
	case PARSER_RULE_CONTEXT_TYPEOF_KEYWORD:
		return &ERROR_MISSING_TYPEOF_KEYWORD
	case PARSER_RULE_CONTEXT_IS_KEYWORD:
		return &ERROR_MISSING_IS_KEYWORD
	case PARSER_RULE_CONTEXT_NULL_KEYWORD:
		return &ERROR_MISSING_NULL_KEYWORD
	case PARSER_RULE_CONTEXT_LOCK_KEYWORD:
		return &ERROR_MISSING_LOCK_KEYWORD
	case PARSER_RULE_CONTEXT_FORK_KEYWORD:
		return &ERROR_MISSING_FORK_KEYWORD
	case PARSER_RULE_CONTEXT_TRAP_KEYWORD:
		return &ERROR_MISSING_TRAP_KEYWORD
	case PARSER_RULE_CONTEXT_IN_KEYWORD:
		return &ERROR_MISSING_IN_KEYWORD
	case PARSER_RULE_CONTEXT_FOREACH_KEYWORD:
		return &ERROR_MISSING_FOREACH_KEYWORD
	case PARSER_RULE_CONTEXT_TABLE_KEYWORD:
		return &ERROR_MISSING_TABLE_KEYWORD
	case PARSER_RULE_CONTEXT_KEY_KEYWORD:
		return &ERROR_MISSING_KEY_KEYWORD
	case PARSER_RULE_CONTEXT_LET_KEYWORD:
		return &ERROR_MISSING_LET_KEYWORD
	case PARSER_RULE_CONTEXT_NEW_KEYWORD:
		return &ERROR_MISSING_NEW_KEYWORD
	case PARSER_RULE_CONTEXT_FROM_KEYWORD:
		return &ERROR_MISSING_FROM_KEYWORD
	case PARSER_RULE_CONTEXT_WHERE_KEYWORD:
		return &ERROR_MISSING_WHERE_KEYWORD
	case PARSER_RULE_CONTEXT_SELECT_KEYWORD:
		return &ERROR_MISSING_SELECT_KEYWORD
	case PARSER_RULE_CONTEXT_START_KEYWORD:
		return &ERROR_MISSING_START_KEYWORD
	case PARSER_RULE_CONTEXT_FLUSH_KEYWORD:
		return &ERROR_MISSING_FLUSH_KEYWORD
	case PARSER_RULE_CONTEXT_WAIT_KEYWORD:
		return &ERROR_MISSING_WAIT_KEYWORD
	case PARSER_RULE_CONTEXT_DO_KEYWORD:
		return &ERROR_MISSING_DO_KEYWORD
	case PARSER_RULE_CONTEXT_TRANSACTION_KEYWORD:
		return &ERROR_MISSING_TRANSACTION_KEYWORD
	case PARSER_RULE_CONTEXT_TRANSACTIONAL_KEYWORD:
		return &ERROR_MISSING_TRANSACTIONAL_KEYWORD
	case PARSER_RULE_CONTEXT_COMMIT_KEYWORD:
		return &ERROR_MISSING_COMMIT_KEYWORD
	case PARSER_RULE_CONTEXT_ROLLBACK_KEYWORD:
		return &ERROR_MISSING_ROLLBACK_KEYWORD
	case PARSER_RULE_CONTEXT_RETRY_KEYWORD:
		return &ERROR_MISSING_RETRY_KEYWORD
	case PARSER_RULE_CONTEXT_ENUM_KEYWORD:
		return &ERROR_MISSING_ENUM_KEYWORD
	case PARSER_RULE_CONTEXT_BASE16_KEYWORD:
		return &ERROR_MISSING_BASE16_KEYWORD
	case PARSER_RULE_CONTEXT_BASE64_KEYWORD:
		return &ERROR_MISSING_BASE64_KEYWORD
	case PARSER_RULE_CONTEXT_MATCH_KEYWORD:
		return &ERROR_MISSING_MATCH_KEYWORD
	case PARSER_RULE_CONTEXT_CONFLICT_KEYWORD:
		return &ERROR_MISSING_CONFLICT_KEYWORD
	case PARSER_RULE_CONTEXT_LIMIT_KEYWORD:
		return &ERROR_MISSING_LIMIT_KEYWORD
	case PARSER_RULE_CONTEXT_ORDER_KEYWORD:
		return &ERROR_MISSING_ORDER_KEYWORD
	case PARSER_RULE_CONTEXT_BY_KEYWORD:
		return &ERROR_MISSING_BY_KEYWORD
	case PARSER_RULE_CONTEXT_GROUP_KEYWORD:
		return &ERROR_MISSING_GROUP_KEYWORD
	case PARSER_RULE_CONTEXT_ORDER_DIRECTION:
		return &ERROR_MISSING_ASCENDING_KEYWORD
	case PARSER_RULE_CONTEXT_JOIN_KEYWORD:
		return &ERROR_MISSING_JOIN_KEYWORD
	case PARSER_RULE_CONTEXT_OUTER_KEYWORD:
		return &ERROR_MISSING_OUTER_KEYWORD
	case PARSER_RULE_CONTEXT_FAIL_KEYWORD:
		return &ERROR_MISSING_FAIL_KEYWORD
	case PARSER_RULE_CONTEXT_PIPE,
		PARSER_RULE_CONTEXT_UNION_OR_INTERSECTION_TOKEN:
		return &ERROR_MISSING_PIPE_TOKEN
	case PARSER_RULE_CONTEXT_EQUALS_KEYWORD:
		return &ERROR_MISSING_EQUALS_KEYWORD
	case PARSER_RULE_CONTEXT_REMOTE_IDENT:
		return &ERROR_MISSING_REMOTE_KEYWORD

	// Type keywords
	case PARSER_RULE_CONTEXT_STRING_KEYWORD:
		return &ERROR_MISSING_STRING_KEYWORD
	case PARSER_RULE_CONTEXT_XML_KEYWORD:
		return &ERROR_MISSING_XML_KEYWORD
	case PARSER_RULE_CONTEXT_RE_KEYWORD:
		return &ERROR_MISSING_RE_KEYWORD
	case PARSER_RULE_CONTEXT_VAR_KEYWORD:
		return &ERROR_MISSING_VAR_KEYWORD
	case PARSER_RULE_CONTEXT_MAP_KEYWORD,
		PARSER_RULE_CONTEXT_NAMED_WORKER_DECL,
		PARSER_RULE_CONTEXT_MAP_TYPE_DESCRIPTOR:
		return &ERROR_MISSING_MAP_KEYWORD
	case PARSER_RULE_CONTEXT_ERROR_KEYWORD,
		PARSER_RULE_CONTEXT_ERROR_BINDING_PATTERN,
		PARSER_RULE_CONTEXT_PARAMETERIZED_TYPE:
		return &ERROR_MISSING_ERROR_KEYWORD
	case PARSER_RULE_CONTEXT_STREAM_KEYWORD:
		return &ERROR_MISSING_STREAM_KEYWORD
	case PARSER_RULE_CONTEXT_READONLY_KEYWORD:
		return &ERROR_MISSING_READONLY_KEYWORD
	case PARSER_RULE_CONTEXT_DISTINCT_KEYWORD:
		return &ERROR_MISSING_DISTINCT_KEYWORD
	case PARSER_RULE_CONTEXT_CLASS_KEYWORD,
		PARSER_RULE_CONTEXT_FIRST_CLASS_TYPE_QUALIFIER,
		PARSER_RULE_CONTEXT_SECOND_CLASS_TYPE_QUALIFIER,
		PARSER_RULE_CONTEXT_THIRD_CLASS_TYPE_QUALIFIER,
		PARSER_RULE_CONTEXT_FOURTH_CLASS_TYPE_QUALIFIER:
		return &ERROR_MISSING_CLASS_KEYWORD
	case PARSER_RULE_CONTEXT_COLLECT_KEYWORD:
		return &ERROR_MISSING_COLLECT_KEYWORD
	case PARSER_RULE_CONTEXT_NATURAL_KEYWORD:
		return &ERROR_MISSING_NATURAL_KEYWORD
	default:
		return &ERROR_SYNTAX_ERROR
	}
}
