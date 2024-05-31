with open('cookies', 'r') as f:
    cookies = f.read()

print_key = ['sessdata', 'bili_jct', 'buvid3', 'dedeuserid', 'ac_time_value']

for cookie_data in cookies.split(';'):
    key, value = cookie_data.split('=', 1)
    key = key.strip().lower()
    value = value.strip()
    if key in print_key:
        print(key, ' = ', f'"{value}"')
