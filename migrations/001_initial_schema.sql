CREATE TABLE IF NOT EXISTS unusual_days (
    id bigserial PRIMARY KEY,
    unusual_date date NOT NULL
);

CREATE TABLE IF NOT EXISTS duties (
    id bigserial PRIMARY KEY,
    duty_id text NOT NULL,
    last_duty_date date
);

CREATE UNIQUE INDEX IF NOT EXISTS duties_duty_id_uidx
    ON duties (duty_id);
