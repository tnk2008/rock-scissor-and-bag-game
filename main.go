package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)
import "github.com/gin-contrib/cors"

// GameResult represents the result of a game round
type GameResult struct {
	PlayerChoice   string `json:"player_choice"`
	ComputerChoice string `json:"computer_choice"`
	Winner         string `json:"winner"`
}

// Database connection
var db *sql.DB

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
func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/game_db")
	if err != nil {
		log.Fatal(err)
	}

	// Ping database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to database")
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

// storeGameRound stores the game round in the database
func storeGameRound(playerChoice, computerChoice, winner string) {
	_, err := db.Exec("INSERT INTO game_rounds (player_choice, computer_choice, winner) VALUES (?, ?, ?)", playerChoice, computerChoice, winner)
	if err != nil {
		log.Fatal(err)
	}
}
