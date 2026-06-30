/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package mcp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultToolName(t *testing.T) {
	require.Equal(t, "get_v1_user_id", defaultToolName("GET", "/v1/user/:id"))
	require.Equal(t, "post_v1_user_id", defaultToolName("POST", "/v1/user/:id"))
	require.Equal(t, "get_static_file_filepath", defaultToolName("GET", "/static/file/*filepath"))
	require.Equal(t, "tool", defaultToolName("", ""))
	require.Equal(t, "get_api_v1_user_list", defaultToolName("GET", "/api/v1/user-list"))
}
