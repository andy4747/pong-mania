-- Drop the trigger
DROP TRIGGER IF EXISTS trigger_set_winner ON scores;

-- Drop the trigger function
DROP FUNCTION IF EXISTS set_winner;

-- Drop the scores table
DROP TABLE IF EXISTS scores;

