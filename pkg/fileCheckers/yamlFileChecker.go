/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package fileCheckers

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/galasa-dev/githubapp-copyright/pkg/checkTypes"
)

type YamlFileChecker struct {
	hashCopyrightPattern         *regexp.Regexp
	hashExpectedCopyrightMessage string
}

func NewYamlFileChecker() FileChecker {
	this := new(YamlFileChecker)

	// \s means any whitespace character (including \n new lines)
	// [*] means a splat/star/asterisk character.
	// We are trying to all this:
	// A copyright message "Copyright contributors to the Galasa project" followed by
	// any number of lines with leading and trailing whitespace around an asterisk, followed by
	// a line containing <optional-whitespace>SPDX-License-Identifier:<optional-whitespace>EPL-2.0
	this.hashCopyrightPattern = regexp.MustCompile(`Copyright contributors to the Galasa project(\s*[#]\s*)*\s*[#]\s*SPDX-License-Identifier:\s*EPL-2[.]0`)
	this.hashExpectedCopyrightMessage = "\nExpected to see:\n#\n# Copyright contributors to the Galasa project\n#\n# SPDX-License-Identifier: EPL-2.0\n#"

	return this
}

func (this *YamlFileChecker) CheckFileContent(content string, fileName string) *checkTypes.CheckError {
	var checkError *checkTypes.CheckError = nil

	commentBlock := ""

	//if it is a bash script (.sh)
	//ignore the first line that starts with !#
	//and any subsequent whitespace
	if strings.HasSuffix(fileName, ".sh") {
		nextLine := strings.Index(content, "\n")
		content = strings.TrimSpace(content[nextLine:])
	}

	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			break
		}
		commentBlock = commentBlock + line + "\n"
	}

	// check we have a comment block at the begining of the file
	if commentBlock == "" {
		checkError = &checkTypes.CheckError{
			Path:     fileName,
			Message:  "A comment block is missing at the start of the file." + this.hashExpectedCopyrightMessage,
			Location: 0,
		}
	} else {
		checkError = checkCommentBlock(&commentBlock, fileName, this.hashCopyrightPattern, this.hashExpectedCopyrightMessage)
	}

	return checkError
}
