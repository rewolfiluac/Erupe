-- Correct the HR5 guide reward that was previously seeded with the wrong items.
-- Official/progression references identify this reception reward as HR5突破の褒賞品;
-- Ferias armor data maps Coord production to item 2496 (decimal 9366, 清雅な装袋).
DO $$
DECLARE
    dist_id INTEGER;
BEGIN
    SELECT id INTO dist_id
    FROM distribution
    WHERE type = 1
      AND character_id IS NULL
      AND event_name IN ('HR5 Welcome Pack', 'HR5突破の褒賞品')
    ORDER BY id
    LIMIT 1;

    IF dist_id IS NULL THEN
        INSERT INTO distribution (type, event_name, description, times_acceptable, min_hr, max_hr)
        VALUES (1, 'HR5突破の褒賞品', '~C05コルーデ装備を生産できる素材を受け取れます。', 1, 5, 999)
        RETURNING id INTO dist_id;
    ELSE
        UPDATE distribution
        SET event_name = 'HR5突破の褒賞品',
            description = '~C05コルーデ装備を生産できる素材を受け取れます。',
            times_acceptable = 1,
            min_hr = 5,
            max_hr = 999
        WHERE id = dist_id;

        DELETE FROM distribution_items WHERE distribution_id = dist_id;
        DELETE FROM distributions_accepted WHERE distribution_id = dist_id;
    END IF;

    INSERT INTO distribution_items (distribution_id, item_type, item_id, quantity)
    SELECT dist_id, 7, 9366, 30
    WHERE NOT EXISTS (
        SELECT 1 FROM distribution_items
        WHERE distribution_id = dist_id
          AND item_type = 7
          AND item_id = 9366
          AND quantity = 30
    );
END $$;
