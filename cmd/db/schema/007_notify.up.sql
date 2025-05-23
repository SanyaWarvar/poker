create table notifications(
    id uuid primary key,
    user_id uuid references users(id),
    payload text,
    last_send_at timestamptz,
    readed bool
);