import random

# Размер игрового поля
SIZE = 4

# Инициализация игрового поля
def initialize_board():
    board = [[0] * SIZE for _ in range(SIZE)]
    add_random_tile(board)
    add_random_tile(board)
    return board

# Добавление случайной плитки (2 или 4) на пустую клетку
def add_random_tile(board):
    empty_cells = [(i, j) for i in range(SIZE) for j in range(SIZE) if board[i][j] == 0]
    if empty_cells:
        i, j = random.choice(empty_cells)
        board[i][j] = random.choice([2, 4])

# Отображение игрового поля
def print_board(board):
    print("-" * (SIZE * 6 + 1))
    for row in board:
        print("|" + "|".join(f"{num:^5}" if num != 0 else "     " for num in row) + "|")
        print("-" * (SIZE * 6 + 1))

# Сдвиг плиток влево
def move_left(board):
    new_board = [[0] * SIZE for _ in range(SIZE)]
    for i in range(SIZE):
        row = [num for num in board[i] if num != 0]
        merged_row = []
        skip = False
        for j in range(len(row)):
            if skip:
                skip = False
                continue
            if j + 1 < len(row) and row[j] == row[j + 1]:
                merged_row.append(row[j] * 2)
                skip = True
            else:
                merged_row.append(row[j])
        merged_row.extend([0] * (SIZE - len(merged_row)))
        new_board[i] = merged_row
    return new_board

# Поворот игрового поля на 90 градусов (для реализации других направлений)
def rotate_board(board):
    return [list(row) for row in zip(*board[::-1])]

# Движение в разных направлениях
def move(board, direction):
    if direction == "a":  # Влево
        return move_left(board)
    elif direction == "d":  # Вправо
        rotated = rotate_board(rotate_board(board))
        moved = move_left(rotated)
        return rotate_board(rotate_board(moved))
    elif direction == "w":  # Вверх
        rotated = rotate_board(rotate_board(rotate_board(board)))
        moved = move_left(rotated)
        return rotate_board(moved)
    elif direction == "s":  # Вниз
        rotated = rotate_board(board)
        moved = move_left(rotated)
        return rotate_board(rotate_board(rotate_board(moved)))
    return board

# Проверка на победу
def check_win(board):
    return any(2048 in row for row in board)

# Проверка на возможность ходов
def can_move(board):
    for i in range(SIZE):
        for j in range(SIZE):
            if board[i][j] == 0:
                return True
            if i + 1 < SIZE and board[i][j] == board[i + 1][j]:
                return True
            if j + 1 < SIZE and board[i][j] == board[i][j + 1]:
                return True
    return False

# Основная функция игры
def play_2048():
    board = initialize_board()
    print("Добро пожаловать в игру 2048!")
    print("Используйте клавиши W (вверх), A (влево), S (вниз), D (вправо) для движения.")
    print("Соберите плитку 2048, чтобы выиграть!")
    while True:
        print_board(board)
        if check_win(board):
            print("Поздравляем! Вы выиграли!")
            break
        if not can_move(board):
            print("Игра окончена. Нет возможных ходов.")
            break
        direction = input("Введите направление (W/A/S/D): ").lower()
        if direction not in ["w", "a", "s", "d"]:
            print("Некорректный ввод. Используйте W, A, S или D.")
            continue
        new_board = move(board, direction)
        if new_board != board:
            board = new_board
            add_random_tile(board)
        else:
            print("Невозможно двигаться в этом направлении!")

# Запуск игры
play_2048()