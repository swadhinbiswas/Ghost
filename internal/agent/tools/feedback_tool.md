Collect and manage user feedback on AI responses. Use this to rate responses, add comments, and view feedback analytics.

**Actions:**
- `rate`: Rate a response (requires `message_id`, `rating`: 1=positive, -1=negative, 0=neutral)
- `comment`: Add a comment to a response (requires `message_id`, optional `comment`, `category`)
- `stats`: View feedback statistics and satisfaction rate
- `insights`: View pattern analysis from feedback
- `list`: View recent feedback entries (optional `limit`, default 10)
- `clear`: Clear all feedback data

**Categories:**
- `accuracy`: Response correctness
- `speed`: Response time
- `tone`: Communication style
- `completeness`: Thoroughness
- `code_quality`: Code output quality

**Examples:**
```json
{"action": "rate", "message_id": "msg_123", "rating": 1, "comment": "Very helpful!", "category": "accuracy"}
{"action": "stats"}
{"action": "insights"}
{"action": "list", "limit": 5}
```
