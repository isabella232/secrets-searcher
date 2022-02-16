package manip_test

import (
	"testing"

	. "github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CodeContext() tests

type codeContextTest struct {
	contents   string
	code       string
	context    string
	wantBefore string
	wantAfter  string
}

func runCodeContextTest(t *testing.T, tt codeContextTest) {
	codeRange := FindLineRange(tt.contents, tt.code)
	contextRange := FindLineRange(tt.contents, tt.context)
	require.NotNil(t, codeRange)
	require.NotNil(t, contextRange)

	// Fire
	gotBefore, gotAfter := CodeContext(tt.contents, codeRange, contextRange, -1)

	require.NotNil(t, gotBefore)
	require.NotNil(t, gotAfter)
	assert.Equal(t, tt.wantBefore, gotBefore.ExtractValue(tt.contents).Value)
	assert.Equal(t, tt.wantAfter, gotAfter.ExtractValue(tt.contents).Value)
}

func TestCodeContext1(t *testing.T) {
	runCodeContextTest(t, codeContextTest{
		contents:   ` "code" `,
		code:       `code`,
		context:    `"code"`,
		wantBefore: `"`,
		wantAfter:  `"`,
	})
}

func TestCodeContext2(t *testing.T) {
	runCodeContextTest(t, codeContextTest{
		contents: `first line
 "code" `,
		code:       `code`,
		context:    `"code"`,
		wantBefore: `"`,
		wantAfter:  `"`,
	})
}

func TestCodeContext3(t *testing.T) {
	runCodeContextTest(t, codeContextTest{
		contents: `first line
 "code breaks
line" `,
		code: `code breaks
line`,
		context: `"code breaks
line"`,
		wantBefore: `"`,
		wantAfter:  `"`,
	})
}

func TestCodeContext4(t *testing.T) {
	runCodeContextTest(t, codeContextTest{
		contents: `first line
 BEFORE-CONTEXT-BREAKS-
LINEcode breaks
lineAFTER-CONTEXT-BREAKS-
LINE `,
		code: `code breaks
line`,
		context: `BEFORE-CONTEXT-BREAKS-
LINEcode breaks
lineAFTER-CONTEXT-BREAKS-
LINE`,
		wantBefore: `BEFORE-CONTEXT-BREAKS-
LINE`,
		wantAfter: `AFTER-CONTEXT-BREAKS-`,
	})
}

//
// CreateCodeContext() tests

type createCodeContextTest struct {
	contents    string
	code        string
	limit       int
	wantContext string
}

func runCreateCodeContextTest(t *testing.T, tt createCodeContextTest) {
	codeRange := FindLineRange(tt.contents, tt.code)
	require.NotNil(t, codeRange)

	// Fire
	contextRange := CreateCodeContext(tt.contents, codeRange, tt.limit)

	require.NotNil(t, contextRange)
	assert.Equal(t, tt.wantContext, contextRange.ExtractValue(tt.contents).Value)
}

func TestCreateCodeContext_1(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents:    ` "code" `,
		code:        `code`,
		limit:       -1,
		wantContext: `"code"`,
	})
}

func TestCreateCodeContext_11(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents:    `define('TENDER_READ_ONLY_API_KEY', 'ebe597e5a240a1663c7587f681158c5031d97b2d');`,
		code:        `API_KEY', 'ebe597e5a240a1663c7587f681158c5031d97b2d'`,
		limit:       -1,
		wantContext: `define('TENDER_READ_ONLY_API_KEY', 'ebe597e5a240a1663c7587f681158c5031d97b2d');`,
	})
}

func TestCreateCodeContext2(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `first line
 "code" `,
		code:        `code`,
		limit:       -1,
		wantContext: `"code"`,
	})
}

func TestCreateCodeContext3(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `first line
 "code breaks
line" `,
		code: `code breaks
line`,
		limit: -1,
		wantContext: `"code breaks
line"`,
	})
}

func TestCreateCodeContext4(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `first line
 "PART-OF
CONTEXT"code breaks
line"PART-OF
CONTEXT" `,
		code: `code breaks
line`,
		limit: -1,
		wantContext: `CONTEXT"code breaks
line"PART-OF`,
	})
}

func TestCreateCodeContext5(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----
`,
		code: `-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----
`,
		limit: -1,
		wantContext: `-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----
`,
	})
}

func TestCreateCodeContext6(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `         '''-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----'''
`,
		code: `-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----`,
		limit: -1,
		wantContext: `'''-----BEGIN CERTIFICATE-----
izfrNTmQLnfsLzi2Wb9xPz2Qj9fQYGgeug3N2MkDuVHwpPcgkhHkJgCQuuvT+qZI
MbS2U6wTS24SZk5RunJIUkitRKeWWMS28SLGfkDs1bBYlSPa5smAd3/q1OePi4ae
-----END CERTIFICATE-----'''`,
	})
}

func TestCreateCodeContext_Limit(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `first line
 BEFORE-CONTEXT-BREAKS-
LINEcode breaks
lineAFTER
CONTEXT-BREAKS-LINE `,
		code: `code breaks
line`,
		limit: 10,
		wantContext: `LINEcode breaks
lineAFTER`,
	})
}

func TestCreateCodeContext_LimitShouldNotIncludeLineBreaks(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `first line
 BEFORE-CONTEXT-BREAKS
LINEcode breaks
lineAFTER
CONTEXT-BREAKS-LINE `,
		code: `code breaks
line`,
		limit: 10,
		wantContext: `LINEcode breaks
lineAFTER`,
	})
}

func TestCreateCodeContext_LimitRealistic(t *testing.T) {
	runCreateCodeContextTest(t, createCodeContextTest{
		contents: `<?php
define('TENDER_DOMAIN', 'pantheon-systems.tenderapp.com');
define('TENDER_SSO_KEY', 'c67fc1cee30f1d3ddf8e22cb9f90c0f495fa86b02b524bff13b83b96280eab1dcdb48801207b6deaea3cdd0e014fe686a5d117c4bfb7adf8f2728d429860ba6c');
define('TENDER_API_KEY', '5c5ba108fb9365482b503a3c3580c941d60e4c88');
define('TENDER_READ_ONLY_API_KEY', 'ebe597e5a240a1663c7587f681158c5031d97b2d');
`,
		code:        `API_KEY', '5c5ba108fb9365482b503a3c3580c941d60e4c88'`,
		limit:       10,
		wantContext: `e('TENDER_API_KEY', '5c5ba108fb9365482b503a3c3580c941d60e4c88');`,
	})
}

func TestCreateCodeContext_Realistic(t *testing.T) {
	//	const contents = `/**
	// * @file
	// * Gigya ratings
	// */
	//(function ($) {
	//    /**
	//     * @todo Undocumented Code!
	//     */
	//    Drupal.gigya = Drupal.gigya || {};
	//    Drupal.gigya.showRatings = function (params) {
	//      gigya.socialize.showRatingUI(params);
	//
	//    };
	//    Drupal.behaviors.gigyaRatings = {
	//      attach: function (context, settings) {
	//        if (typeof gigya !== 'undefined') {
	//          if (typeof Drupal.settings.gigyaRaitingsInstances !== 'undefined') {
	//            $.each(Drupal.settings.gigyaRaitingsInstances, function (index, rating) {
	//              Drupal.gigya.showRatings(rating);
	//            });
	//
	//          }
	//        }
	//      }
	//    };
	//})(jQuery);
	//
	//`
	//	runCreateCodeContextTest(t, createCodeContextTest{
	//		contents:    contents,
	//		code:        NewLineRange(0, 675).ExtractValue(contents).Value,
	//		limit:       50,
	//		wantContext: ,
	//	})
}
