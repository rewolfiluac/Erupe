-- Remove non-official demo distribution rewards from existing databases.
DELETE FROM distributions_accepted
WHERE distribution_id IN (
    SELECT id FROM distribution
    WHERE type = 1
      AND event_name IN ('Extra Item Storage', 'Extra Equipment Storage')
);

DELETE FROM distribution_items
WHERE distribution_id IN (
    SELECT id FROM distribution
    WHERE type = 1
      AND event_name IN ('Extra Item Storage', 'Extra Equipment Storage')
);

DELETE FROM distribution
WHERE type = 1
  AND event_name IN ('Extra Item Storage', 'Extra Equipment Storage');
