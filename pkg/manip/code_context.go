package manip

import "regexp"

func CodeContext(contents string, codeRange, contextRangeInput *LineRange, limit int) (before, after *LineRange) {

	// We'll create our own context range, since we can use the end index
	// If we don't have a contextRange passed, we'll use the start too.
	generatedContextRange := CreateCodeContext(contents, codeRange, limit)

	var contextRange *LineRange
	if contextRangeInput == nil {
		contextRange = generatedContextRange
	} else {
		contextRange = NewLineRange(contextRangeInput.StartIndex, generatedContextRange.EndIndex)
	}

	if contextRange != nil && codeRange.StartIndex < contextRange.StartIndex {
		panic("context must start with or before code")
	}
	if codeRange.EndIndex > contextRange.EndIndex {
		panic("context must end with or after code")
	}

	before = NewLineRange(contextRange.StartIndex, codeRange.StartIndex)
	after = NewLineRange(codeRange.EndIndex, contextRange.EndIndex)

	return
}

func CreateCodeContext(contents string, codeRange *LineRange, limit int) (result *LineRange) {
	// Get everything before code
	allBeforeRange := NewLineRange(0, codeRange.StartIndex)
	var beforeOffset int
	if limit > -1 && allBeforeRange.Len() > limit {
		beforeOffset = allBeforeRange.Len() - limit
		allBeforeRange = NewLineRange(allBeforeRange.EndIndex-limit, allBeforeRange.EndIndex)
	}
	allBeforeCode := allBeforeRange.ExtractValue(contents).Value

	// Get everything after code
	allAfterRange := NewLineRange(codeRange.EndIndex, len(contents))
	if limit > 1 && allAfterRange.Len() > limit {
		allAfterRange = NewLineRange(allAfterRange.StartIndex, allAfterRange.StartIndex+limit)
	}
	allAfterCode := allAfterRange.ExtractValue(contents).Value

	// Get non-whitespace characters before code to return
	reBefore := regexp.MustCompile(`[^\s]*$`)
	beforeMatch := reBefore.FindStringIndex(allBeforeCode)
	if beforeMatch == nil {
		panic("This expression should match anything")
	}
	contextStartIndex := beforeOffset + beforeMatch[0]

	// Get non-whitespace characters after code to return
	reAfter := regexp.MustCompile(`^[^\s]*`)
	afterMatch := reAfter.FindStringSubmatchIndex(allAfterCode)
	if afterMatch == nil {
		panic("This expression should match anything")
	}
	contextEndIndex := codeRange.EndIndex + afterMatch[1]

	result = NewLineRange(contextStartIndex, contextEndIndex)

	return result
}
