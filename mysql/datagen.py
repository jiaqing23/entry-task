 
import hashlib

def md5(str2hash: str) -> str:
    result = hashlib.md5(str2hash.encode())
    return result.hexdigest()

table_name = 'user'
number = 100 * 1000
sql = "INSERT INTO %s VALUES " % table_name

for i in range(number):
    user = "user-" + str(i)
    password = md5(user)
    sql += "('%s','%s','%s')," % (user, user, password)

# Change the last "," to ";"
sql = sql[:-1] + ";"

sql_file = open("insert.sql", "w")
sql_file.write(sql)
sql_file.close()
