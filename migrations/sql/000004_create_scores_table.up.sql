CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    player1_id INTEGER NOT NULL,
    player2_id INTEGER NOT NULL,
    player1_score INTEGER NOT NULL,
    player2_score INTEGER NOT NULL,
    room_id VARCHAR(255),
    game_ended_at TIMESTAMP NOT NULL,
    winner INTEGER,
    FOREIGN KEY (player1_id) REFERENCES users(id),
    FOREIGN KEY (player2_id) REFERENCES users(id)
);

-- function to calculate the winner
CREATE OR REPLACE FUNCTION set_game_winner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.player1_score > NEW.player2_score THEN
        NEW.winner = NEW.player1_id;
    ELSEIF NEW.player2_score > NEW.player1_score THEN
        NEW.winner = NEW.player2_id;
    ELSE
        NEW.winner = NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- trigger to set the winner automatically
CREATE TRIGGER trigger_set_game_winner
BEFORE INSERT OR UPDATE ON scores
FOR EACH ROW
    EXECUTE FUNCTION set_game_winner();

