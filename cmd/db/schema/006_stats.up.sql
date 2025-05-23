create table player_stats(
    user_id uuid not null references users(id) primary key,
    games_played int default 0 not null,
    max_balance int default 1000 not null
);