DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM duties
        WHERE last_duty_date IS NOT NULL
        GROUP BY last_duty_date
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION
            'cannot enforce one duty per date: duplicate last_duty_date values exist';
    END IF;
END
$$;

CREATE UNIQUE INDEX IF NOT EXISTS duties_last_duty_date_uidx
    ON duties (last_duty_date)
    WHERE last_duty_date IS NOT NULL;
