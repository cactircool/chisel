#include <cstdlib>
#include <ios>
#include <iosfwd>
#include <streambuf>
#include <istream>
#include <vector>
#include <algorithm>
#include <sstream>

namespace chisel {

	class CountingStreamBuffer : public std::streambuf {
		std::streambuf *_underlying;
		std::vector<std::streampos> _lines;
		size_t _line_index = 0;
		size_t _column = 0;
		std::streampos _current_pos = 0;
		bool _last_was_cr = false; // for \r\n

	public:
		CountingStreamBuffer(std::streambuf *underlying) : _underlying(underlying) {
			_lines.push_back(0);
		}

		size_t line() const noexcept { return _line_index; }
		size_t column() const noexcept { return _column; }

		std::streampos line_start(size_t line) const {
			if (line < _lines.size())
				return _lines[line];
			return std::streampos(-1);
		}

		size_t total_lines() const { return _lines.size(); }
	protected:
		int_type underflow() override {
			return _underlying->sgetc();
		}

		int_type uflow() override {
			int_type ch = _underlying->sbumpc();
			if (ch == traits_type::eof()) return traits_type::eof();

			process_character(static_cast<char_type>(ch));
			return ch;
		}

		std::streamsize xsgetn(char_type *s, std::streamsize n) override {
			std::streamsize count = _underlying->sgetn(s, n);
			for (std::streamsize i = 0; i < count; ++i)
				process_character(s[i]);
			return count;
		}

		std::streampos seekoff(std::streamoff off, std::ios_base::seekdir dir, std::ios_base::openmode which) override {
			std::streampos new_pos = _underlying->pubseekoff(off, dir, which);
			if (new_pos != std::streampos(-1))
				update_position_after_seek(new_pos);
			return new_pos;
		}

		std::streampos seekpos(std::streampos pos, std::ios_base::openmode which) override {
			std::streampos new_pos = _underlying->pubseekpos(pos, which);
			if (new_pos != std::streampos(-1))
				update_position_after_seek(new_pos);
			return new_pos;
		}

		std::streamsize showmanyc() override {
			return _underlying->in_avail();
		}

		int_type pbackfail(int_type ch = traits_type::eof()) override {
			int_type result = _underlying->sputbackc(traits_type::to_char_type(ch));
			if (result != traits_type::eof())
				if (_current_pos > 0) {
					_current_pos += -1;
					update_position_after_seek(_current_pos);
				}
			return result;
		}
	private:
		void process_character(char_type ch) {
			_current_pos += 1;
			if (ch == '\n') {
				if (!_last_was_cr)
					start_new_line();
				_last_was_cr = false;
			} else if (ch == '\r') {
				start_new_line();
				_last_was_cr = true;
			} else {
				++_column;
				_last_was_cr = false;
			}
		}

		void start_new_line() {
			++_line_index;
			_column = 0;
			if (_line_index >= _lines.size())
				_lines.push_back(_current_pos);
		}

		void update_position_after_seek(std::streampos pos) {
			_current_pos = pos;
			_last_was_cr = false;

			if (pos == std::streampos(0)) {
				_line_index = 0;
				_column = 0;
				return;
			}

			if (pos <= _current_pos) {
				_current_pos = pos;
				auto it = std::upper_bound(_lines.begin(), _lines.end(), pos);
				if (it == _lines.begin()) {
					_line_index = 0;
					_column = static_cast<size_t>(pos);
				} else {
					--it;
					_line_index = std::distance(_lines.begin(), it);
					_column = static_cast<size_t>(pos - *it);
				}
				return;
			}

			std::streampos target_pos = pos;
			std::streampos original_pos = _current_pos;

			std::streampos underlying_pos = _underlying->pubseekoff(0, std::ios_base::cur, std::ios_base::in);
			_underlying->pubseekpos(original_pos, std::ios_base::in);

			while (_current_pos < target_pos) {
				int_type ch = _underlying->sbumpc();
				if (ch == traits_type::eof()) {
					_current_pos = _underlying->pubseekoff(0, std::ios_base::cur, std::ios_base::in);
					break;
				}
				process_character(static_cast<char_type>(ch));
			}

			_underlying->pubseekpos(pos, std::ios_base::in);
		}
	};

	class Reader : public std::istream {
		CountingStreamBuffer *_buffer;
	public:
		explicit Reader(std::istream &reader) : std::istream(nullptr), _buffer(new CountingStreamBuffer(reader.rdbuf())) {
			rdbuf(_buffer);
			copyfmt(reader);
			clear(reader.rdstate());
		}

		Reader(const Reader &) = delete;
		Reader &operator=(const Reader &) = delete;

		~Reader() {
			delete _buffer;
		}

		size_t line() const {
			return _buffer->line() + 1;
		}
		size_t column() const {
			return _buffer->column() + 1;
		}

		std::streampos line_start(size_t line) const {
			return _buffer->line_start(line);
		}
		size_t total_lines() const noexcept {
			return _buffer->total_lines();
		}

		CountingStreamBuffer *buffer() noexcept {
			return _buffer;
		}
		const CountingStreamBuffer *buffer() const noexcept {
			return _buffer;
		}

		std::string error(const std::string &msg) const {
			std::stringstream ss;
			ss << line() << ":" << column() << ' ' << msg << std::endl;
			return ss.str();
		}
	};

}
