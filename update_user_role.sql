-- Update the user with the provided email to have instructor role
UPDATE users SET role = 'instructor' WHERE email = 'test@example.com';

-- Verify the update
SELECT id, name, email, role FROM users WHERE email = 'test@example.com'; 