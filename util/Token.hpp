#include <cstring>
#include <algorithm>
#include <cstdint>
#include <vector>

namespace chisel {

	struct Token {
		enum class Type {
			{{.TokenTypes}}
		};

	private:
		Type _type;
		TOKEN_UTILE_TYPE *_data;
		TOKEN_LENGTH_TYPE _len;

		static std::vector<TOKEN_UTILE_TYPE *> _strings;

		constexpr static const TOKEN_UTILE_TYPE *_static_data[] = {
			{{.StaticTypeValues}}
		};
		constexpr static const int _static_lens[] = {
			{{.StaticTypeLengths}}
		};

		constexpr static const char *_type_names[] = {
			{{.TypeNames}}
		};

		static bool matches(const TOKEN_UTILE_TYPE *a, TOKEN_LENGTH_TYPE alen, const TOKEN_UTILE_TYPE *b, TOKEN_LENGTH_TYPE blen) {
	        if (!a || !b || alen != blen) return false;
	        return memcmp(a, b, alen) == 0;
	    }

		static constexpr TOKEN_UTILE_TYPE FAILED_SENTINEL = 0;
		static char _failed;

	public:
		Token() : _type(), _data(&_failed), _len(0) {}
		Token(Type type) : _type(type), _data(nullptr), _len(0) {}
		Token(Type type, TOKEN_UTILE_TYPE *data, TOKEN_LENGTH_TYPE len) : _type(type), _data(data), _len(len) {}

		template <typename T>
		Token(Type type, TOKEN_UTILE_TYPE *data, T len) : _type(type), _data(data), _len(static_cast<TOKEN_LENGTH_TYPE>(len)) {}

		Token(const Token &other) : _type(other._type), _len(other._len) {
			if (other._data && other._data != &_failed) {
	            _data = new char[_len + 1];
	            memcpy(_data, other._data, _len);
	            _data[_len] = '\0';
	        } else {
	            _data = other._data;
	        }
		}
		Token &operator=(const Token &other) {
			if (this != &other) {
	            // Clean up old data
	            if (_data && _data != &_failed) {
	                delete[] _data;
	            }

	            _type = other._type;
	            _len = other._len;

	            if (other._data && other._data != &_failed) {
	                _data = new char[_len + 1];
	                memcpy(_data, other._data, _len);
	                _data[_len] = '\0';
	            } else {
	                _data = other._data;
	            }
	        }
	        return *this;
		}

		Token(Token &&other) noexcept : _type(other._type), _data(other._data), _len(other._len) {
			other._data = nullptr;
			other._len = 0;
		}
		Token &operator=(Token &&other) noexcept {
			if (this != &other) {
	            // Clean up old data
	            if (_data && _data != &_failed) {
	                delete[] _data;
	            }

	            _type = other._type;
	            _data = other._data;
	            _len = other._len;

	            other._data = nullptr;
	            other._len = 0;
	        }
	        return *this;
		}

		~Token() {
			if (_data && _data != &_failed)
				delete[] _data;
		}

		Type type() const {
			return _type;
		}

		TOKEN_LENGTH_TYPE length() const {
			if (_data && _data != &_failed) return _len;
        	return _static_lens[static_cast<unsigned int>(_type)];
		}
		TOKEN_LENGTH_TYPE len() const {
			return length();
		}
		TOKEN_LENGTH_TYPE size() const {
			return length();
		}

		const TOKEN_UTILE_TYPE *static_data() const {
			return _data;
		}
		const TOKEN_UTILE_TYPE *data() const {
			if (_data == &_failed) return nullptr;
			if (!_data) return _static_data[static_cast<unsigned int>(_type)];
			return _data;
		}

		friend bool operator==(const Token &a, Type b) {
			return a.type() == b;
		}
		friend bool operator==(Type a, const Token &b) {
			return a == b.type();
		}

		friend bool operator==(const Token &a, const TOKEN_UTILE_TYPE *b) {
			return (a.data() == b) || matches(a.static_data(), a.len(), b, TOKEN_DATA_STRLEN(b));
		}
		friend bool operator==(const TOKEN_UTILE_TYPE *a, const Token &b) {
			return (b.data() == a) || matches(b.static_data(), b.len(), a, TOKEN_DATA_STRLEN(a));
		}

		bool operator==(const Token &other) const {
			return type() == other.type() && (data() == other.data() || matches(static_data(), len(), other.static_data(), other.len()));
		}

		template <typename T>
		friend bool operator!=(const Token &a, T b) {
			return !(a == b);
		}

		template <typename T>
		friend bool operator!=(T a, const Token &b) {
			return !(a == b);
		}

		static const char *name(Type type) {
			return _type_names[static_cast<unsigned int>(type)];
		}

		static Token failed;

		operator bool() const {
			return _data != &_failed;
		}
	};

	TOKEN_UTILE_TYPE Token::_failed = Token::FAILED_SENTINEL;
	Token Token::failed = Token();

}
