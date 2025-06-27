conn = new Mongo();
db = conn.getDB(process.env.MONGO_INITDB_DATABASE);


db.createCollection("hash_crack_subtasks", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["_id", "taskId", "partNumber", "data", "percent", "status", "createdAt", "updatedAt"],
            properties: {
                _id: {
                    bsonType: "objectId",
                    description: "Уникальный идентификатор подзадачи"
                },
                taskId: {
                    bsonType: "objectId",
                    description: "ID основной задачи"
                },
                partNumber: {
                    bsonType: "int",
                    description: "Номер части задачи",
                    minimum: 0
                },
                data: {
                    bsonType: "array",
                    items: {
                        bsonType: "string"
                    },
                    description: "Данные для перебора"
                },
                percent: {
                    bsonType: "double",
                    description: "Процент выполнения",
                    minimum: 0,
                    maximum: 100
                },
                status: {
                    enum: ["PENDING", "IN_PROGRESS", "SUCCESS", "ERROR", "UNKNOWN"],
                    description: "Статус выполнения подзадачи"
                },
                reason: {
                    bsonType: "string",
                    description: "Причина ошибки (если есть)"
                },
                createdAt: {
                    bsonType: "date",
                    description: "Время создания подзадачи"
                },
                updatedAt: {
                    bsonType: "date",
                    description: "Время последнего обновления задачи"
                }
            }
        }
    }
});


db.hash_crack_subtasks.createIndex({taskId: 1, partNumber: 1}, {unique: true});


db.hash_crack_subtasks.createIndex({status: 1});


db.hash_crack_subtasks.createIndex({createdAt: 1});


db.createCollection("hash_crack_tasks", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["_id", "hash", "maxLength", "partCount", "status", "createdAt", "updatedAt"],
            properties: {
                _id: {
                    bsonType: "objectId",
                    description: "Уникальный идентификатор задачи"
                },
                hash: {
                    bsonType: "string",
                    description: "Хеш для подбора"
                },
                maxLength: {
                    bsonType: "int",
                    description: "Максимальная длина пароля",
                    minimum: 1
                },
                partCount: {
                    bsonType: "int",
                    description: "Количество частей",
                    minimum: 1
                },
                status: {
                    enum: ["PENDING", "IN_PROGRESS", "PARTIAL_READY", "READY", "ERROR", "UNKNOWN"],
                    description: "Статус выполнения задачи"
                },
                reason: {
                    bsonType: "string",
                    description: "Причина ошибки (если есть)"
                },
                finishedAt: {
                    bsonType: "date",
                    description: "Время завершения задачи (если завершена)"
                },
                createdAt: {
                    bsonType: "date",
                    description: "Время создания задачи"
                },
                updatedAt: {
                    bsonType: "date",
                    description: "Время последнего обновления задачи"
                }
            }
        }
    }
});


db.hash_crack_tasks.createIndex({hash: 1, maxLength: 1});


db.hash_crack_tasks.createIndex({status: 1});


db.hash_crack_tasks.createIndex({createdAt: 1});


db.createView("hash_crack_tasks_with_subtasks", "hash_crack_tasks", [
    {
        $lookup: {
            from: "hash_crack_subtasks",
            localField: "_id",
            foreignField: "taskId",
            as: "subtasks"
        }
    }
]);