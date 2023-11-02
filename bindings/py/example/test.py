import multiprocessing
import minify  # type: ignore


def minify_worker(number: int) -> None:
    minify.string('text/html', f"<p>{number}</p>"*10000)
    print(".", end="", flush=True)


if __name__ == "__main__":

    number_list = range(0, 1000)

    processing_pool = multiprocessing.Pool()
    processing_pool.map(minify_worker, number_list)

    print("Done!")
