package kube

// func TestKubernetesClient_RetrieveIngresses(t *testing.T) {
// 	client := fake.NewSimpleClientset()
// 	ctx := context.Background()

// 	type fields struct {
// 		client *fake.Clientset
// 	}
// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   map[string]bool
// 	}{
// 		{
// 			name: "Test Ingress",
// 			fields: fields{
// 				client: client,
// 			},
// 			args: args{
// 				ctx: ctx,
// 			},
// 			want: map[string]bool{
// 				"foo/bar": true,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			k := KubernetesClient{
// 				client: tt.fields.client,
// 			}
// 			if got := k.RetrieveIngresses(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("KubernetesClient.RetrieveIngresses() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
