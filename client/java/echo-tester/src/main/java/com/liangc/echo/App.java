package com.liangc.echo;

import jnr.unixsocket.UnixSocketAddress;
import jnr.unixsocket.UnixSocketChannel;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.SocketChannel;
import java.util.Optional;
import java.util.concurrent.FutureTask;
import java.util.function.Consumer;

/**
 * Hello AF-ECHO!
 */
public class App {

    private static String flagNet(String[] args) {
        try {
            return args[args.length - 2];
        } catch (Exception e) {
        }
        return null;
    }

    private static Integer flagTotal(String[] args) {
        try {
            return Integer.parseInt(args[args.length - 1]);
        } catch (Exception e) {
        }
        return null;
    }

    public static void main(String[] args) throws Exception {
        System.out.println(System.currentTimeMillis());
        App app = new App();
        String msg = "hello world.";
        Integer total = Optional.ofNullable(flagTotal(args)).orElse(100000);
        String net = Optional.ofNullable(flagNet(args)).orElse("unix");
        long startTime = 0;
        switch (net) {
            case "unix":
                startTime = app.unix("/tmp/afecho.ipc", msg, total);
                break;
            case "tcp":
                startTime = app.tcp("localhost", 12345, msg, total);
                break;
        }
        long finalTime = System.currentTimeMillis() - startTime;
        System.out.printf("done : net=%s , msg.size=%d , total=%d , time=%dms , avg=%f/s\r\n",
                net, msg.getBytes().length, total, finalTime, total.floatValue() / (finalTime / 1000.0));
    }

    long unix(String fp, String msg, int total) throws Exception {
        UnixSocketAddress address = new UnixSocketAddress(fp);
        UnixSocketChannel channel = UnixSocketChannel.open(address);
        long s = System.currentTimeMillis();
        echoHandler(channel, msg, total);
        return s;
    }

    long tcp(String host, int port, String msg, int total) throws Exception {
        SocketChannel channel = SocketChannel.open();
        channel.connect(new InetSocketAddress(host, port));
        long s = System.currentTimeMillis();
        echoHandler(channel, msg, total);
        return s;
    }

    void echoHandler(SocketChannel channel, String msg, int total) throws Exception {
        FutureTask<Consumer<SocketChannel>> ft = new FutureTask<>(() -> {
            try {
                int t = 0;
                while (t < total) {
                    ByteBuffer readBuffer = ByteBuffer.allocate(128);
                    channel.read(readBuffer);
                    StringBuilder stringBuffer = new StringBuilder();
                    readBuffer.flip();
                    while (readBuffer.hasRemaining()) {
                        stringBuffer.append((char) readBuffer.get());
                    }
                    t += stringBuffer.toString().split("\n").length;
                }
            } catch (Exception e) {
                e.printStackTrace();
            }
        }, ch -> {
            try {
                ch.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        });
        new Thread(ft).start();
        for (int i = 0; i < total; i++) {
            ByteBuffer writeBuffer = ByteBuffer.allocate(128);
            writeBuffer.put((msg + "\n").getBytes());
            writeBuffer.flip();
            channel.write(writeBuffer);
        }
        ft.get().accept(channel);
    }

}


