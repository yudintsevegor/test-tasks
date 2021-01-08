SELECT
  result.Date_Time AS day, result.name AS department_name, SUM(result.Hours) AS total_hours
FROM (
  SELECT
 department.name AS Name, department.id, timesheet.department_id, DATE_PART('hour', timesheet.logout::timestamp) - DATE_PART('hour', timesheet.login::timestamp) AS Hours, timesheet.login::date AS Date_Time
FROM
  department, timesheet
WHERE
  department.id = timesheet.department_id
ORDER BY
  Date_Time, department.name
) AS result
GROUP BY day, department_name