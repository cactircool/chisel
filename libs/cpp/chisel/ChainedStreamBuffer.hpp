#ifndef CHISEL_CHAINED_STREAM_BUFFER_HPP
#define CHISEL_CHAINED_STREAM_BUFFER_HPP

#include <iostream>
#include <streambuf>
#include <list>

namespace chisel {

	class ChainedStreamBuffer : public std::streambuf {
	public:
		struct Buffer {
			std::streambuf *buffer;
			bool deletable;

			~Buffer() {
				if (deletable)
					delete buffer;
			}
		};

	    explicit ChainedStreamBuffer(const std::list<Buffer> &bufs)
	        : _bufs(bufs) {}

		void prepend_buffer(std::streambuf *buffer) {
			_bufs.insert(_bufs.begin(), { .buffer = buffer, .deletable = true });
		}

		void append_buffer(std::streambuf *buffer) {
			_bufs.insert(_bufs.end(), { .buffer = buffer, .deletable = true });
		}

	protected:
	    int underflow() override {
			while (!_bufs.empty()) {
				int ch = _bufs.front().buffer->sgetc();
				if (ch == traits_type::eof()) {
					_bufs.pop_front();
					continue;
				}
				return ch;
			}
	        return traits_type::eof();
	    }

	    int uflow() override {
      		while (!_bufs.empty()) {
				int ch = _bufs.front().buffer->sbumpc();
				if (ch == traits_type::eof()) {
					_bufs.pop_front();
					continue;
				}
				return ch;
			}
	        return traits_type::eof();
	    }

		int pbackfail(int c) override {
            // Try to put back character to current buffer
            if (!_bufs.empty() && _bufs.front().buffer->sputbackc(c) != traits_type::eof())
                return c;
            return traits_type::eof(); // Putback failed
        }

	private:
	    std::list<Buffer> _bufs;
	};

}

#endif
