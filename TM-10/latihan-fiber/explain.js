db = db.getSiblingDB("praktikum_backend");

print("=== EXPLAIN 1: Query student_id (Index: portfolio_student_unique) ===");
const exp1 = db.student_portfolios.find({ student_id: "std_001" }).explain("executionStats");
print("stage: " + exp1.executionStats.executionStages.stage);
print("nReturned: " + exp1.executionStats.nReturned);
print("totalDocsExamined: " + exp1.executionStats.totalDocsExamined);
print("totalKeysExamined: " + exp1.executionStats.totalKeysExamined);
printjson(exp1.executionStats.executionStages);

print("\n=== EXPLAIN 2: Query projects.technologies (Index: portfolio_technologies - Multikey) ===");
const exp2 = db.student_portfolios.find({ "projects.technologies": "go" }).explain("executionStats");
print("stage: " + exp2.executionStats.executionStages.stage);
print("nReturned: " + exp2.executionStats.nReturned);
print("totalDocsExamined: " + exp2.executionStats.totalDocsExamined);
print("totalKeysExamined: " + exp2.executionStats.totalKeysExamined);
printjson(exp2.executionStats.executionStages);
