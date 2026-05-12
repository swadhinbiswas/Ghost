package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const feedbackFile = "feedback.json"

// Rating represents a user's rating of an AI response.
type Rating int

const (
	RatingNeutral  Rating = 0
	RatingPositive Rating = 1
	RatingNegative Rating = -1
)

// Feedback represents a single piece of user feedback.
type Feedback struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	MessageID string    `json:"message_id"`
	Rating    Rating    `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	Category  string    `json:"category,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Context   string    `json:"context,omitempty"`
}

// Store manages persistent feedback storage.
type Store struct {
	mu       sync.RWMutex
	feedback []*Feedback
	dataPath string
}

// NewStore creates a new feedback store.
func NewStore(workingDir string) (*Store, error) {
	s := &Store{
		dataPath: filepath.Join(workingDir, ".ghost", feedbackFile),
	}
	return s, s.Load()
}

// Load reads feedback from disk.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.feedback = make([]*Feedback, 0)
			return nil
		}
		return fmt.Errorf("failed to read feedback file: %w", err)
	}

	if err := json.Unmarshal(data, &s.feedback); err != nil {
		s.feedback = make([]*Feedback, 0)
		return nil
	}

	return nil
}

// Save writes feedback to disk.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.feedback, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataPath, data, 0644)
}

// Add adds a new feedback entry and persists it.
func (s *Store) Add(f *Feedback) error {
	s.mu.Lock()
	f.ID = fmt.Sprintf("fb-%d", len(s.feedback)+1)
	f.Timestamp = time.Now()
	s.feedback = append(s.feedback, f)
	s.mu.Unlock()

	return s.Save()
}

// GetBySession returns all feedback for a session.
func (s *Store) GetBySession(sessionID string) []*Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Feedback
	for _, f := range s.feedback {
		if f.SessionID == sessionID {
			result = append(result, f)
		}
	}
	return result
}

// GetByMessage returns feedback for a specific message.
func (s *Store) GetByMessage(messageID string) *Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.feedback {
		if f.MessageID == messageID {
			return f
		}
	}
	return nil
}

// GetAll returns all feedback entries.
func (s *Store) GetAll() []*Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Feedback, len(s.feedback))
	copy(result, s.feedback)
	return result
}

// GetStats returns feedback statistics.
func (s *Store) GetStats() FeedbackStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := FeedbackStats{}
	for _, f := range s.feedback {
		stats.Total++
		switch f.Rating {
		case RatingPositive:
			stats.Positive++
		case RatingNegative:
			stats.Negative++
		default:
			stats.Neutral++
		}

		if f.Category != "" {
			if stats.ByCategory == nil {
				stats.ByCategory = make(map[string]int)
			}
			stats.ByCategory[f.Category]++
		}
	}

	if stats.Total > 0 {
		stats.SatisfactionRate = float64(stats.Positive) / float64(stats.Total) * 100
	}

	return stats
}

// Delete removes a feedback entry by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, f := range s.feedback {
		if f.ID == id {
			s.feedback = append(s.feedback[:i], s.feedback[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("feedback not found: %s", id)
}

// Clear removes all feedback.
func (s *Store) Clear() error {
	s.mu.Lock()
	s.feedback = make([]*Feedback, 0)
	s.mu.Unlock()
	return s.Save()
}

// GetRecent returns the N most recent feedback entries.
func (s *Store) GetRecent(n int) []*Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n >= len(s.feedback) {
		result := make([]*Feedback, len(s.feedback))
		copy(result, s.feedback)
		return result
	}

	start := len(s.feedback) - n
	result := make([]*Feedback, n)
	copy(result, s.feedback[start:])
	return result
}

// GetPatternInsights analyzes feedback to find patterns and suggestions.
func (s *Store) GetPatternInsights() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var insights []string

	// Analyze negative feedback patterns
	negativeComments := make(map[string]int)
	for _, f := range s.feedback {
		if f.Rating == RatingNegative && f.Comment != "" {
			// Extract keywords from comments
			words := extractKeywords(f.Comment)
			for _, word := range words {
				negativeComments[word]++
			}
		}
	}

	// Report common issues
	for word, count := range negativeComments {
		if count >= 2 {
			insights = append(insights, fmt.Sprintf("Frequent issue: '%s' mentioned in %d negative feedbacks", word, count))
		}
	}

	// Analyze category trends
	if len(s.feedback) > 0 {
		categoryCounts := make(map[string]int)
		categoryRatings := make(map[string][]Rating)
		for _, f := range s.feedback {
			if f.Category != "" {
				categoryCounts[f.Category]++
				categoryRatings[f.Category] = append(categoryRatings[f.Category], f.Rating)
			}
		}

		for cat, ratings := range categoryRatings {
			if len(ratings) >= 3 {
				avgRating := 0.0
				for _, r := range ratings {
					avgRating += float64(r)
				}
				avgRating /= float64(len(ratings))

				if avgRating < 0 {
					insights = append(insights, fmt.Sprintf("Low satisfaction in category '%s' (avg rating: %.2f)", cat, avgRating))
				}
			}
		}
	}

	return insights
}

// FeedbackStats holds aggregated feedback statistics.
type FeedbackStats struct {
	Total            int            `json:"total"`
	Positive         int            `json:"positive"`
	Negative         int            `json:"negative"`
	Neutral          int            `json:"neutral"`
	SatisfactionRate float64        `json:"satisfaction_rate"`
	ByCategory       map[string]int `json:"by_category,omitempty"`
}

// extractKeywords extracts meaningful keywords from a comment.
func extractKeywords(comment string) []string {
	// Simple keyword extraction - could be enhanced with NLP
	skipWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "it": true, "to": true, "of": true,
		"and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "for": true, "with": true, "by": true, "from": true,
		"this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "we": true,
		"they": true, "me": true, "him": true, "her": true, "us": true,
		"them": true, "my": true, "your": true, "his": true, "its": true,
		"our": true, "their": true, "not": true, "no": true, "very": true,
		"just": true, "like": true, "would": true, "could": true, "should": true,
	}

	words := make(map[string]bool)
	current := ""
	for _, r := range comment {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			current += string(r)
		} else {
			if len(current) > 3 {
				lower := toLower(current)
				if !skipWords[lower] {
					words[lower] = true
				}
			}
			current = ""
		}
	}
	// Don't forget the last word
	if len(current) > 3 {
		lower := toLower(current)
		if !skipWords[lower] {
			words[lower] = true
		}
	}

	var result []string
	for w := range words {
		result = append(result, w)
	}
	return result
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = byte(r + 32)
		} else {
			result[i] = byte(r)
		}
	}
	return string(result)
}
