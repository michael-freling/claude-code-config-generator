package workflow

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

const ClaudeResumeFlag = "--resume"

// SessionManager handles Claude CLI session lifecycle
type SessionManager struct {
	logger Logger
}

// SessionInfo holds information about a Claude session
type SessionInfo struct {
	SessionID  string
	CreatedAt  time.Time
	ReuseCount int
	IsNew      bool // Whether this is a new session or reused
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger Logger) *SessionManager {
	return &SessionManager{
		logger: logger,
	}
}

func (m *SessionManager) ParseSessionID(output string) string {
	if output == "" {
		return ""
	}

	// Parse each line as JSON (stream-json format)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as JSON
		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			// Not valid JSON, skip
			continue
		}

		// Check for session_id in different locations
		if sessionID := m.extractSessionIDFromChunk(chunk); sessionID != "" {
			if m.logger != nil {
				m.logger.Verbose("Found session ID: %s", sessionID)
			}
			return sessionID
		}
	}

	// Fallback: use regex to find session_id patterns
	sessionID := m.extractSessionIDWithRegex(output)
	if sessionID != "" && m.logger != nil {
		m.logger.Verbose("Found session ID with regex: %s", sessionID)
	}

	return sessionID
}

func (m *SessionManager) extractSessionIDFromChunk(chunk map[string]interface{}) string {
	chunkType, _ := chunk["type"].(string)

	// Check in result type chunks
	if chunkType == "result" {
		if sessionID, ok := chunk["session_id"].(string); ok && sessionID != "" {
			return sessionID
		}
	}

	// Check in system chunks with init subtype
	if chunkType == "system" {
		if subtype, ok := chunk["subtype"].(string); ok && subtype == "init" {
			if sessionID, ok := chunk["session_id"].(string); ok && sessionID != "" {
				return sessionID
			}
		}
	}

	return ""
}

func (m *SessionManager) extractSessionIDWithRegex(output string) string {
	// Pattern 1: "session_id":"..."
	re1 := regexp.MustCompile(`"session_id"\s*:\s*"([^"]+)"`)
	if matches := re1.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: session_id: ... (without quotes)
	re2 := regexp.MustCompile(`session_id\s*:\s*([a-zA-Z0-9\-]+)`)
	if matches := re2.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// BuildCommandArgs adds --resume flag if sessionID is provided and not forcing new session
func (m *SessionManager) BuildCommandArgs(sessionID string, forceNewSession bool) []string {
	if forceNewSession || sessionID == "" {
		return nil
	}

	if m.logger != nil {
		m.logger.Verbose("Reusing Claude session: %s", sessionID)
	}

	return []string{ClaudeResumeFlag, sessionID}
}

// GetSessionFromState returns nil if no session exists in state
func (m *SessionManager) GetSessionFromState(state *WorkflowState) *SessionInfo {
	if state == nil || state.SessionID == nil || *state.SessionID == "" {
		return nil
	}

	info := &SessionInfo{
		SessionID:  *state.SessionID,
		ReuseCount: state.SessionReuseCount,
		IsNew:      false,
	}

	if state.SessionCreatedAt != nil {
		info.CreatedAt = *state.SessionCreatedAt
	}

	return info
}

// UpdateStateWithSession increments reuse count if session already exists
func (m *SessionManager) UpdateStateWithSession(state *WorkflowState, sessionID string, isNew bool) {
	if state == nil || sessionID == "" {
		return
	}

	if isNew {
		// New session
		state.SessionID = &sessionID
		now := time.Now()
		state.SessionCreatedAt = &now
		state.SessionReuseCount = 0

		if m.logger != nil {
			m.logger.Verbose("Created new Claude session: %s", sessionID)
		}
	} else {
		// Reusing existing session
		state.SessionID = &sessionID
		state.SessionReuseCount++

		if m.logger != nil {
			m.logger.Verbose("Reused Claude session: %s (count: %d)", sessionID, state.SessionReuseCount)
		}
	}
}
