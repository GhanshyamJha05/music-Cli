package online

// Genre defines a browsable music category with a display name,
// emoji, colour hint, and the yt-dlp search query used to fetch picks.
type Genre struct {
	Name  string
	Emoji string
	Color string // hex for lipgloss
	Query string // sent to yt-dlp
}

// Genres is the curated list shown on the Recommended screen.
var Genres = []Genre{
	{
		Name:  "Pop",
		Emoji: "🎤",
		Color: "#FF6EB4",
		Query: "best pop songs 2024 official audio",
	},
	{
		Name:  "Lo-Fi",
		Emoji: "☕",
		Color: "#A78BFA",
		Query: "lofi hip hop chill beats study relax",
	},
	{
		Name:  "Romantic",
		Emoji: "💕",
		Color: "#FB7185",
		Query: "best romantic love songs playlist",
	},
	{
		Name:  "Funk",
		Emoji: "🎸",
		Color: "#F59E0B",
		Query: "best funk songs groove playlist",
	},
	{
		Name:  "Rock",
		Emoji: "🤘",
		Color: "#EF4444",
		Query: "best rock songs classic hits playlist",
	},
	{
		Name:  "Hip-Hop",
		Emoji: "🎧",
		Color: "#8B5CF6",
		Query: "best hip hop rap songs 2024",
	},
	{
		Name:  "Jazz",
		Emoji: "🎷",
		Color: "#14B8A6",
		Query: "smooth jazz piano saxophone playlist",
	},
	{
		Name:  "Electronic",
		Emoji: "🎛️",
		Color: "#06B6D4",
		Query: "best electronic dance music EDM playlist",
	},
	{
		Name:  "Classical",
		Emoji: "🎻",
		Color: "#D4A574",
		Query: "best classical music orchestral pieces",
	},
	{
		Name:  "R&B / Soul",
		Emoji: "🎵",
		Color: "#F97316",
		Query: "best rnb soul music playlist 2024",
	},
	{
		Name:  "Country",
		Emoji: "🤠",
		Color: "#84CC16",
		Query: "best country music songs playlist",
	},
	{
		Name:  "Bollywood",
		Emoji: "🎬",
		Color: "#EC4899",
		Query: "best bollywood songs hindi music hits",
	},
}

// RecommendedSearch fetches yt-dlp results for a genre query.
func RecommendedSearch(genre Genre) ([]OnlineTrack, error) {
	return YtdlpSearch(genre.Query, 15)
}
