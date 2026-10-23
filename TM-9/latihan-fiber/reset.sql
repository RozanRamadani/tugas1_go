DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);

-- Reset supervisions untuk D.2 / D.3
DELETE FROM supervisions;

UPDATE supervisors
SET supervised_count = (SELECT COUNT(*) FROM supervisions WHERE supervisor_id = supervisors.id);
