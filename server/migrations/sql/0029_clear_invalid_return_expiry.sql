-- New and recently active users should not have an active return-player
-- expiry. Earlier registration paths initialized return_expires to 30 days
-- in the future, which made normal users look eligible for return worlds.
UPDATE users
   SET return_expires = NULL
 WHERE return_expires > now()
   AND (
       last_login IS NULL
       OR last_login > now() - interval '90 days'
   );
