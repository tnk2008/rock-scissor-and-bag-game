package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/microsoft/go-mssqldb"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// GameResult represents the result of a game round
type GameResult struct {
	PlayerChoice   string `json:"player_choice"`
	ComputerChoice string `json:"computer_choice"`
	Winner         string `json:"winner"`
}

// Database connection
var db *sql.DB
var server = "tna.database.windows.net"
var port = 1433
var user = "tna"
var password = "Collins123"
var database = "gamedb"

func main() {

	r := gin.Default()
	// Add CORS middleware
	r.Use(cors.Default())

	// Initialize database connection
	initDB()

	// Define API endpoints
	r.POST("/play", playGame)
	r.GET("/stats", getGameStatistics)
	r.GET("/allGames", getAllGames)

	// Run HTTP server
	r.Run(":8080")
}
func getAllGames(c *gin.Context) {
	var gameRounds []GameResult
	rows, err := db.Query("SELECT player_choice, computer_choice, winner FROM game_rounds")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var gameRound GameResult
		err := rows.Scan(&gameRound.PlayerChoice, &gameRound.ComputerChoice, &gameRound.Winner)
		if err != nil {
			log.Fatal(err)
		}
		gameRounds = append(gameRounds, gameRound)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK, gameRounds)

}

// playGame is the handler for playing the game
func playGame(c *gin.Context) {
	playerChoice := c.PostForm("choice")

	// Validate input choice
	if playerChoice != "rock" && playerChoice != "scissors" && playerChoice != "bag" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid choice. Must be one of: 'rock', 'scissors', 'bag'"})
		return
	}

	// Determine computer's choice
	computerChoice := getComputerChoice()

	// Determine winner
	winner := determineWinner(playerChoice, computerChoice)

	// Store game round in database
	storeGameRound(playerChoice, computerChoice, winner)

	c.JSON(http.StatusOK, GameResult{PlayerChoice: playerChoice, ComputerChoice: computerChoice, Winner: winner})
}

// getGameStatistics is the handler for retrieving game statistics
func getGameStatistics(c *gin.Context) {
	var totalGames int
	var totalWins int

	// Retrieve total number of game rounds
	err := db.QueryRow("SELECT COUNT(*) FROM game_rounds").Scan(&totalGames)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve total number of wins
	err = db.QueryRow("SELECT COUNT(*) FROM game_rounds WHERE winner = 'player'").Scan(&totalWins)
	if err != nil {
		log.Fatal(err)
	}

	// Calculate win percentage
	winPercentage := float64(totalWins) / float64(totalGames) * 100

	c.JSON(http.StatusOK, gin.H{"total_games": totalGames, "total_wins": totalWins, "win_percentage": winPercentage})
}

// initDB initializes the database connection
// initDB initializes the database connection
func initDB() {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)
	var err error
	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Connected!")
}

// getComputerChoice generates a random choice for the computer
func getComputerChoice() string {
	choices := []string{"rock", "scissors", "bag"}
	rand.Seed(time.Now().UnixNano())
	return choices[rand.Intn(len(choices))]
}

// determineWinner determines the winner of the game round
func determineWinner(playerChoice, computerChoice string) string {
	if (playerChoice == "rock" && computerChoice == "scissors") ||
		(playerChoice == "scissors" && computerChoice == "bag") ||
		(playerChoice == "bag" && computerChoice == "rock") {
		return "player"
	} else if playerChoice == computerChoice {
		return "tie"
	} else {
		return "computer"
	}
}
var ctx = context.Background()
// storeGameRound stores the game round in the database
func storeGameRound(playerChoice, computerChoice, winner string) {
	query := `
        INSERT INTO game_rounds (player_choice, computer_choice, winner, created_at)
        VALUES (@playerChoice, @computerChoice, @winner, GETDATE())
    `
	_, err := db.ExecContext(ctx, query,
		sql.Named("playerChoice", playerChoice),
		sql.Named("computerChoice", computerChoice),
		sql.Named("winner", winner),
	)
	if err != nil {
		log.Fatal(err)
	}
}
